package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/viper"
	"github.com/wandermaia/desafio-temperatura-cep/service-a/internal/infra/webserver/handlers"
)

// load env vars cfg
func init() {
	//viper.AutomaticEnv()
	// viper.SetDefault("TITLE", "Microservice Demo")
	// viper.SetDefault("BACKGROUND_COLOR", "green")
	// viper.SetDefault("RESPONSE_TIME", "1000")
	// viper.SetDefault("EXTERNAL_CALL_URL", "http://goapp2:8181")
	//http://service-b:8282/
	//viper.SetDefault("SERVICE_B_URL", "http://service-b:8282/")
	// viper.SetDefault("EXTERNAL_CALL_METHOD", "GET")
	// viper.SetDefault("REQUEST_NAME_OTEL", "microservice-demo")
	// viper.SetDefault("OTEL_SERVICE_NAME", "microservice-demo")
	// viper.SetDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "otel-collector:4317")
	viper.SetDefault("HTTP_PORT", ":8181")
}

func main() {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Post("/cep", handlers.BuscaTemperaturaHandler)

	log.Println("Starting server on port", viper.GetString("HTTP_PORT"))
	http.ListenAndServe(viper.GetString("HTTP_PORT"), router)
}

// go mod init github.com/wandermaia/desafio-opentelemetry
// docker rm -f $(docker ps -a -q)
