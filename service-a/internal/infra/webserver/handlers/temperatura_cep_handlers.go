package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/viper"
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

func init() {
	//viper.AutomaticEnv()
	// viper.SetDefault("TITLE", "Microservice Demo")
	// viper.SetDefault("BACKGROUND_COLOR", "green")
	// viper.SetDefault("RESPONSE_TIME", "1000")
	// viper.SetDefault("EXTERNAL_CALL_URL", "http://goapp2:8181")
	//http://service-b:8282/
	viper.SetDefault("SERVICE_B_URL", "http://service-b:8282/")
	// viper.SetDefault("EXTERNAL_CALL_METHOD", "GET")
	// viper.SetDefault("REQUEST_NAME_OTEL", "microservice-demo")
	// viper.SetDefault("OTEL_SERVICE_NAME", "microservice-demo")
	// viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	//viper.SetDefault("HTTP_PORT", ":8181")
	//ExternalCallURL:    viper.GetString("SERVICE_B_URL"),
}

// Função que busca a temperatura no service-b
func BuscaTemperaturaHandler(w http.ResponseWriter, r *http.Request) {

	//Coletando o CEP  partir do body da requisição
	var cepParam DadosCep
	err := json.NewDecoder(r.Body).Decode(&cepParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Caso o cep não esteja em um formato válido, retora o código 422 e a mensagem de erro.
	if !validarFormatoCEP(cepParam.Cep) {
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

	// Realizando a coleta no service-b
	//url := "http://service-b:8282/" + cepParam.Cep
	url := viper.GetString("SERVICE_B_URL") + cepParam.Cep
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Erro chamar a url %s: %s", url, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Erro ao ler a resposta: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Caso o cep esteja em um formato válido, mas não seja encontrado
	if resp.Status == "404 Not Found" {
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
		log.Printf("Erro ao fazer Unmarshal do JSON service-b: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Retornando a resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

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
