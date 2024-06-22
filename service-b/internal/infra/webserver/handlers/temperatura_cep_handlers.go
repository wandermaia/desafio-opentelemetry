package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi"
)

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

// Struct que será utilizada para formar a resposta com o valor das temperaturas
type TemperaturaCidade struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
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

func BuscaTemperaturaHandler(w http.ResponseWriter, r *http.Request) {

	//Coletando o CEP  partir do parâmetro da URL
	cepParam := chi.URLParam(r, "cep")

	// Caso o cep não esteja em um formato válido, retora o código 422 e a mensagem de erro.
	if !validarFormatoCEP(cepParam) {
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

	// Buscando os dados da cidade
	dadosCep, err := BuscaCepViaCep(cepParam)
	if err != nil {
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

	// Coletando a temperatura da cidade
	temperatura, err := ConsultaTemperaturaCidade(dadosCep.Localidade)
	if err != nil {
		log.Printf("Erro ao consultar os parâmetros para a localidade %s: %s", dadosCep.Localidade, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Retornando a resposta
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(temperatura)

}

// Função que vai realizar a consulta dos dados de temperatura da cidade
func ConsultaTemperaturaCidade(cidade string) (*TemperaturaCidade, error) {

	// Variável temporária. O token será alterado
	token := "6ceb0269ea6049eda52220700241706"

	// Realizando o encode para caracteres especiais e espaço
	encodedCidade := url.QueryEscape(cidade)

	// Coletando os daodos no webservice
	url := "http://api.weatherapi.com/v1/current.json?q=" + encodedCidade + "&lang=pt&country=Brazil&key=" + token
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
	var temperatura TemperaturaCidade
	var data ResponseBody
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Printf("Erro ao fazer Unmarshal do JSON weatherapi: %s", err)
		return nil, err
	}

	// Segregando os dados e calculando a temperatura em kelvin a partir da temperatura em Celsius
	temperatura.TempC = data.Current.TempC
	temperatura.TempF = data.Current.TempF
	temperatura.TempK = data.Current.TempC + 273.0

	// Enviando a resposta
	return &temperatura, nil

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
