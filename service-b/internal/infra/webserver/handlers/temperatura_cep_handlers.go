package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

// Struct que será utilizada para receber o cep do path da requisição
type DadosCep struct {
	Cep string `json:"cep"`
}

// Struct para receber os dados para o webserver. A função BuscaTemperaturaHandler está anexada nessa struct. Com isso, ela terá acesso aos dados.
type Webserver struct {
	OtelData *TemplateOtelData
}

// Função que cria um novo webserver com base nos dados informados.
func NewServer(templateOtelData *TemplateOtelData) *Webserver {
	return &Webserver{
		OtelData: templateOtelData,
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
	router.Get("/{cep}", we.BuscaTemperaturaHandler)
	return router
}

// Struct para armazenamento dos dados do OTEL.
type TemplateOtelData struct {
	RequestNameOTEL string
	OTELTracer      trace.Tracer
}

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
	Erro        bool   `json:"erro"`
}

type ResponseBody struct {
	Location struct {
		Name      string  `json:"name"`
		Region    string  `json:"region"`
		Country   string  `json:"country"`
		Lat       float64 `json:"lat"`
		Lon       float64 `json:"lon"`
		TzID      string  `json:"tz_id"`
		Localtime string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int     `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
	} `json:"current"`
}

// Função que busca a temperatura
func (h *Webserver) BuscaTemperaturaHandler(w http.ResponseWriter, r *http.Request) {

	// Carregandos o header para gerar o request id para conseguir a rastreabilidade
	carrier := propagation.HeaderCarrier(r.Header)
	ctx := r.Context()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)

	// Criação de span inicial
	ctx, span := h.OtelData.OTELTracer.Start(ctx, "Início Processamento "+h.OtelData.RequestNameOTEL)
	defer span.End()

	//Coletando o CEP  partir do parâmetro da URL
	cepParam := chi.URLParam(r, "cep")

	// Criação de um span de validação CEP
	ctx, spanCEP := h.OtelData.OTELTracer.Start(ctx, "Validar Formatação CEP")

	// Caso o cep não esteja em um formato válido, retora o código 422 e a mensagem de erro.
	if !validarFormatoCEP(cepParam) {
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

	ctx, spanBuscaCepViaCep := h.OtelData.OTELTracer.Start(ctx, "Busca CEP")
	// Buscando os dados da cidade
	dadosCep, err := BuscaCepViaCep(cepParam)
	if err != nil {
		spanBuscaCepViaCep.SetStatus(codes.Error, "can not find zipcode")
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

	spanBuscaCepViaCep.End()

	ctx, spanConsultaTemperaturaCidade := h.OtelData.OTELTracer.Start(ctx, "Busca Temperatura")

	// Coletando a temperatura da cidade
	climaCidade, err := ConsultaTemperaturaCidade(dadosCep.Localidade)
	if err != nil {
		spanConsultaTemperaturaCidade.SetStatus(codes.Error, "Erro ao consultar os parâmetros para a localidade.")
		log.Printf("Erro ao consultar os parâmetros para a localidade %s: %s", dadosCep.Localidade, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	spanConsultaTemperaturaCidade.End()

	// Retornando a resposta
	ctx, spanEnviandoResposta := h.OtelData.OTELTracer.Start(ctx, "Enviando resposta")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(climaCidade)
	spanEnviandoResposta.End()

}

// Função que vai realizar a consulta dos dados de temperatura da cidade
func ConsultaTemperaturaCidade(cidade string) (*ClimaCidade, error) {

	// Constante para acesso
	const CONSTANTE = "6ceb0269ea6049eda52220700241706"

	// Realizando o encode para caracteres especiais e espaço
	encodedCidade := url.QueryEscape(cidade)

	// Coletando os daodos no webservice
	url := "http://api.weatherapi.com/v1/current.json?q=" + encodedCidade + "&lang=pt&country=Brazil&key=" + CONSTANTE
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Realizando o Unmarshal
	var clima ClimaCidade
	var data ResponseBody
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Erro ao fazer Unmarshal do JSON weatherapi: %s", err)
		return nil, err
	}

	// Segregando os dados e calculando a temperatura em kelvin a partir da temperatura em Celsius
	clima.Cidade = cidade
	clima.TempC = data.Current.TempC
	clima.TempF = data.Current.TempF
	clima.TempK = data.Current.TempC + 273.0

	// Enviando a resposta
	return &clima, nil

}

// Função que realiza a busca no site ViaCep o CEP informado por parâmetro.
func BuscaCepViaCep(cep string) (*ViaCEP, error) {

	url := "http://viacep.com.br/ws/" + cep + "/json/"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var dadosCep ViaCEP
	err = json.Unmarshal(body, &dadosCep)
	if err != nil {
		return nil, err
	}

	// Caso o cep não tenha sido encontrado, a variável "erro" recebe o valor true.
	if dadosCep.Erro {
		return nil, errors.New("can not find zipcode")
	}
	return &dadosCep, nil

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
