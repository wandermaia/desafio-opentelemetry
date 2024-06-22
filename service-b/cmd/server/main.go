package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/wandermaia/desafio-temperatura-cep/service-b/internal/infra/webserver/handlers"
)

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Get("/{cep}", handlers.BuscaTemperaturaHandler)

	log.Println("Servidor iniciado na porta 8282!")
	http.ListenAndServe(":8282", router)
}

// go mod init github.com/wandermaia/desafio-opentelemetry
// docker rm -f $(docker ps -a -q)
