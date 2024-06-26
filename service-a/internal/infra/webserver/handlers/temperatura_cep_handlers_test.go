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
func TestBuscaTemperaturaHandler(t *testing.T) {

	// Server mock para simular o service-b
	serverMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cep" {
			t.Errorf("Expected to request '/cep', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got: %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{ "city: "cidade", "temp_C": 28.5, "temp_F": 28.5, "temp_K": 28.5 }`))
	}))
	defer serverMock.Close()

	// criação do trace.
	tracer := otel.Tracer("microservice-tracer-mock")

	// Dados para a criação do servidor
	templateData := &TemplateData{
		ExternalCallURL: serverMock.URL,
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	// Criação do server
	server := NewServer(templateData)
	router := server.CreateServer()

	//Realizando a chamada
	req, _ := http.NewRequest("POST", "/32450000", nil)
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

// // Cep com formato válido, mas não encontrado. Deve retornar Código 404
// // e a mensagem "can not find zipcode"
func TestBuscaTemperaturaHandlerCepNaoEncontrado(t *testing.T) {

	// Server mock para simular o service-b
	serverMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cep" {
			t.Errorf("Expected to request '/cep', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got: %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{ "message": "can not find zipcode"}`))
	}))
	defer serverMock.Close()

	// criação do trace.
	tracer := otel.Tracer("microservice-tracer-mock")

	// Dados para a criação do servidor
	templateData := &TemplateData{
		ExternalCallURL: serverMock.URL,
		RequestNameOTEL: "microservice-tracer-mock",
		OTELTracer:      tracer,
	}

	// Criação do server
	server := NewServer(templateData)
	router := server.CreateServer()

	//Realizando a chamada
	req, _ := http.NewRequest("POST", "/00000000", nil)
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

//https://medium.com/zus-health/mocking-outbound-http-requests-in-go-youre-probably-doing-it-wrong-60373a38d2aa
