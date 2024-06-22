package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/wandermaia/desafio-opentelemetry/service-b/internal/infra/webserver/handlers"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/{cep}", handlers.BuscaTemperaturaHandler)

	log.Println("Servidor iniciado na porta 8080!")
	http.ListenAndServe(":8080", router)
}

// go mod init github.com/wandermaia/desafio-opentelemetry
