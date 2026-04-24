package mux

import (
	"net/http"

	_ "go-far/api/openapi"

	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

// InitHttpMux initializes the HTTP ServeMux
func InitHttpMux(log *zerolog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Serve swagger.json and swagger.yaml as static files
	swaggerFS := http.Dir("./api/openapi")
	mux.Handle("GET /swagger/swagger.json", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"), // relative URL to your JSON
	))

	return mux
}
