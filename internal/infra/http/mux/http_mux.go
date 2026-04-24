package mux

import (
	"net/http"

	_ "go-far/api/openapi"

	"github.com/arl/statsviz"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

// InitHttpMux initializes the HTTP ServeMux
func InitHttpMux(log *zerolog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Register metrics endpoint for StatsViz
	errStatviz := statsviz.Register(mux)
	if errStatviz != nil {
		log.Error().Err(errStatviz).Msg("Failed to register statsviz handler")
	}

	// // Serve swagger files
	// swaggerFS := http.Dir("./api/openapi")
	// mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))

	// Serve swagger.json and swagger.yaml as static files
	swaggerFS := http.Dir("./api/openapi")

	// Only serve the JSON/YAML files, not a directory listing
	// mux.Handle("GET /swagger/swagger.json", http.FileServer(swaggerFS))
	// mux.Handle("GET /swagger/swagger.yaml", http.FileServer(swaggerFS))
	mux.Handle("GET /swagger/swagger.json", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))

	// Serve Swagger UI (interactive documentation)
	// Tell the UI where to fetch the spec – this must be a URL the browser can GET
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"), // relative URL to your JSON
	))
	// mux.Handle("GET /docs/", httpSwagger.Handler(
	// 	httpSwagger.URL("/swagger/swagger.json"),
	// ))

	return mux
}
