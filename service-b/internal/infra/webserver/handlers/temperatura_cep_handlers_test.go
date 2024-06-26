package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

// Cep Válido. Deve retornar Código 200 e o Response Body
// no formato: { "city: "São Paulo", "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }
func TestBuscaTemperaturaHandlerOk(t *testing.T) {
	// criação do trace.
	tracer := otel.Tracer("microservice-tracer-mock")

	// Dados para a criação do servidor
	templateData := &TemplateOtelData{
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	// Criação do server
	server := NewServer(templateData)
	router := server.CreateServer()

	//Realizando a chamada
	req, _ := http.NewRequest("GET", "/32450000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var clima ClimaCidade
	err := json.Unmarshal(w.Body.Bytes(), &clima)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, clima)
}

// Cep INVÁLIDO (com formato incorreto). Deve retornar Código 422
// e a mensagem "invalid zipcode"
func TestBuscaTemperaturaHandlerCepInvalido(t *testing.T) {
	// criação do trace.
	tracer := otel.Tracer("microservice-tracer-mock")

	// Dados para a criação do servidor
	templateData := &TemplateOtelData{
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	// Criação do server
	server := NewServer(templateData)
	router := server.CreateServer()

	//Realizando a chamada
	req, _ := http.NewRequest("GET", "/324500000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var clima ClimaCidade
	err := json.Unmarshal(w.Body.Bytes(), &clima)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Empty(t, clima)
}

// Cep com formato válido, mas não encontrado. Deve retornar Código 404
// e a mensagem "can not find zipcode"
func TestBuscaTemperaturaHandlerCepNaoEncontrado(t *testing.T) {
	// criação do trace.
	tracer := otel.Tracer("microservice-tracer-mock")

	// Dados para a criação do servidor
	templateData := &TemplateOtelData{
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	// Criação do server
	server := NewServer(templateData)
	router := server.CreateServer()

	//Realizando a chamada
	req, _ := http.NewRequest("GET", "/00000000", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var clima ClimaCidade
	err := json.Unmarshal(w.Body.Bytes(), &clima)
	if err != nil {
		return
	}
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Empty(t, clima)
}
