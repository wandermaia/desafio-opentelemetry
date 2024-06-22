package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/wandermaia/desafio-temperatura-cep/service-a/internal/infra/webserver/handlers"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/{cep}", handlers.BuscaTemperaturaHandler)

	log.Println("Servidor iniciado na porta 8181!")
	http.ListenAndServe(":8181", router)
}

// go mod init github.com/wandermaia/desafio-opentelemetry
// docker rm -f $(docker ps -a -q)