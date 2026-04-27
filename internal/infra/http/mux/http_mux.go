package mux

import (
	"net/http"

	"go-far/api/openapi"
	"go-far/internal/infra/metrics"

	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
)

// InitHttpMux initializes the HTTP ServeMux
func InitHttpMux(log *zerolog.Logger, metricsInst metrics.Metrics) *http.ServeMux {
	mux := http.NewServeMux()

	swaggerFS := http.Dir("./api/openapi")
	mux.Handle("GET /swagger/swagger.json", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
	))

	if metricsInst != nil {
		mux.Handle("/metrics", metricsInst.HTTPHandler())
		mux.Handle("/debug/vars", metricsInst.HTTPHandler())
	}

	_ = openapi.SwaggerInfo

	return mux
}
