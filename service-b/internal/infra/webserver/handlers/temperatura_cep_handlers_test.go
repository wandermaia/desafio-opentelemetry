package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

// Cep Válido. Deve retornar Código 200 e o Response Body
// no formato: { "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }
func TestBuscaTemperaturaHandlerOk(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/{cep}", BuscaTemperaturaHandler)
	req, _ := http.NewRequest("GET", "/32450000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var temperatura TemperaturaCidade
	err := json.Unmarshal(w.Body.Bytes(), &temperatura)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, temperatura)
}

// Cep Válido, mas com caracter especial e espaço (São Paulo) Deve retornar Código 200 e o Response Body
// no formato: { "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }
func TestBuscaTemperaturaHandlerOkCaractereEspecial(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/{cep}", BuscaTemperaturaHandler)
	req, _ := http.NewRequest("GET", "/01021200", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var temperatura TemperaturaCidade
	err := json.Unmarshal(w.Body.Bytes(), &temperatura)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, temperatura)
}

// Cep INVÁLIDO (com formato incorreto). Deve retornar Código 422
// e a mensagem "invalid zipcode"
func TestBuscaTemperaturaHandlerCepInvalido(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/{cep}", BuscaTemperaturaHandler)
	req, _ := http.NewRequest("GET", "/324500000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var temperatura TemperaturaCidade
	err := json.Unmarshal(w.Body.Bytes(), &temperatura)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Empty(t, temperatura)
}

// Cep com formato válido, mas não encontrado. Deve retornar Código 404
// e a mensagem "can not find zipcode"
func TestBuscaTemperaturaHandlerCepNaoEncontrado(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/{cep}", BuscaTemperaturaHandler)
	req, _ := http.NewRequest("GET", "/00000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var temperatura TemperaturaCidade
	err := json.Unmarshal(w.Body.Bytes(), &temperatura)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, temperatura)
}
