package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Struct que será utilizada para formar a resposta com o valor das temperaturas
type ClimaCidade struct {
	Cidade string  `json:"city"`
	TempC  float64 `json:"temp_C"`
	TempF  float64 `json:"temp_F"`
	TempK  float64 `json:"temp_K"`
}

// Struct que será utilizada para receber o cep do body da requisição
type DadosCep struct {
	Cep string `json:"cep"`
}

// Struct para receber os dados para o webserver. A função BuscaTemperaturaHandler está anexada nessa struct. Com isso, ela terá acesso aos dados.
type Webserver struct {
	TemplateData *TemplateData
}

// Função que cria um novo webserver com base nos dados informados.
func NewServer(templateData *TemplateData) *Webserver {
	return &Webserver{
		TemplateData: templateData,
	}
}

// Cria um novo server utilizando o chi e acrescentando alguns midlewares importantes.
func (we *Webserver) CreateServer() *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Logger)
	router.Use(middleware.Timeout(60 * time.Second))
	// promhttp. Usado para gerar as métricas automáticas do prometheus
	router.Handle("/metrics", promhttp.Handler())
	router.Post("/cep", we.BuscaTemperaturaHandler)
	return router
}

// Struct para armazenamento dos dados do webserver e handler.
type TemplateData struct {
	ExternalCallURL string
	RequestNameOTEL string
	OTELTracer      trace.Tracer
}

// func init() {
// 	//viper.AutomaticEnv()
// 	// viper.SetDefault("TITLE", "Microservice Demo")
// 	// viper.SetDefault("BACKGROUND_COLOR", "green")
// 	// viper.SetDefault("RESPONSE_TIME", "1000")
// 	// viper.SetDefault("EXTERNAL_CALL_URL", "http://goapp2:8181")
// 	//http://service-b:8282/
// 	viper.SetDefault("SERVICE_B_URL", "http://service-b:8282/")
// }

// Função que busca a temperatura no service-b
func (h *Webserver) BuscaTemperaturaHandler(w http.ResponseWriter, r *http.Request) {

	// Carregandos o header para gerar o request id para conseguir a rastreabilidade
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Criação de span inicial
	ctx, span := h.TemplateData.OTELTracer.Start(ctx, "Início Processamento "+h.TemplateData.RequestNameOTEL)
	defer span.End()

	// Criação de um span de validação CEP
	ctx, spanCEP := h.TemplateData.OTELTracer.Start(ctx, "Formatação CEP")

	//Coletando o CEP  partir do body da requisição
	var cepParam DadosCep
	err := json.NewDecoder(r.Body).Decode(&cepParam)
	if err != nil {
		spanCEP.SetStatus(codes.Error, "Erro Realizar decode do CEP")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Caso o cep não esteja em um formato válido, retora o código 422 e a mensagem de erro.
	if !validarFormatoCEP(cepParam.Cep) {
		spanCEP.SetStatus(codes.Error, "invalid zipcode")
		log.Printf("invalid zipcode: %s", cepParam)
		msg := struct {
			Message string `json:"message"`
		}{
			Message: "invalid zipcode",
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(msg)
		return

	}

	spanCEP.End()

	// Criação de um span de trace do service-b
	ctx, spanServiceB := h.TemplateData.OTELTracer.Start(ctx, "Consulta service-b")

	// Preparando a URL para a realização da request
	url := h.TemplateData.ExternalCallURL + cepParam.Cep
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		spanServiceB.SetStatus(codes.Error, "invalid url")
		log.Printf("Erro formar a request para a url %s: %s", url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Injetando o header do request id. Necessário para realizar o tracker
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Executando a request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		spanServiceB.SetStatus(codes.Error, "error acces url")
		log.Printf("Erro chamar a url %s: %s", url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler a resposta: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Finalização do span de cosulta ao service-b
	spanServiceB.End()

	// Caso o cep esteja em um formato válido, mas não seja encontrado
	if resp.Status == "404 Not Found" {
		span.SetStatus(codes.Error, "can not find zipcode")
		log.Printf("can not find zipcode: %s", cepParam)
		msg := struct {
			Message string `json:"message"`
		}{
			Message: "can not find zipcode",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(msg)
		return
	}

	// Realizando o Unmarshal
	var clima ClimaCidade
	err = json.Unmarshal(body, &clima)
	if err != nil {
		span.SetStatus(codes.Error, "Erro ao fazer Unmarshal do JSON service-b")
		log.Printf("Erro ao fazer Unmarshal do JSON service-b: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Retornando a resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	span.SetStatus(codes.Ok, "CEP consultado")

	json.NewEncoder(w).Encode(clima)

}

// Função que valida o formato CEP informado por parâmetro
func validarFormatoCEP(parametro string) bool {
	// Verifica se o parâmetro tem exatamente 8 caracteres
	if len(parametro) != 8 {
		return false
	}

	// Verifica se todos os caracteres são números inteiros
	_, err := strconv.Atoi(parametro)
	return err == nil
}
