package query

import (
	"text/template"

	"github.com/rs/zerolog"
)

type QueriesOptions struct {
	Path string `yaml:"path"`
}

type QueryLoader struct {
	log       zerolog.Logger
	templates map[string]*template.Template // key: query name
	rawSQL    map[string]string             // key: query name
	fileMap   map[string]string             // query name -> file path (for error reporting)
}

func InitQueryLoader(log zerolog.Logger, opt QueriesOptions) *QueryLoader {
	ql := &QueryLoader{
		log:       log,
		templates: make(map[string]*template.Template),
		rawSQL:    make(map[string]string),
		fileMap:   make(map[string]string),
	}
	if err := ql.load(opt.Path); err != nil {
		log.Panic().Err(err).Msg("Failed to load queries")
	}

	log.Debug().Msgf("Queries loaded successfully, total queries: %d", len(ql.templates))
	return ql
}
