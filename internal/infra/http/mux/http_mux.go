package mux

import (
	"net/http"

	"github.com/arl/statsviz"
	"github.com/rs/zerolog"
)

// InitHttpMux initializes the HTTP ServeMux
func InitHttpMux(log *zerolog.Logger) *http.ServeMux {
	mux := http.NewServeMux()

	// Register metrics endpoint for StatsViz
	errStatviz := statsviz.Register(mux)
	if errStatviz != nil {
		log.Error().Err(errStatviz).Msg("Failed to register statsviz handler")
	}

	// Serve swagger files
	swaggerFS := http.Dir("./etc/docs")
	mux.Handle("GET /swagger/", http.StripPrefix("/swagger/", http.FileServer(swaggerFS)))

	return mux
}
