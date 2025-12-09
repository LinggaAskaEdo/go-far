package config

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

type QueriesOptions struct {
	Path string `yaml:"path"`
}

type QueryLoader struct {
	queries  map[string]string
	filePath string
}

func InitQueryLoader(log zerolog.Logger, opt QueriesOptions) (*QueryLoader, error) {
	queryPath := fmt.Sprintf("%s/user_queries.sql", opt.Path)
	ql := &QueryLoader{
		queries:  make(map[string]string),
		filePath: queryPath,
	}

	if err := ql.load(log); err != nil {
		return nil, err
	}

	return ql, nil
}

func (ql *QueryLoader) load(log zerolog.Logger) error {
	data, err := os.ReadFile(ql.filePath)
	if err != nil {
		return err
	}

	content := string(data)
	sections := strings.Split(content, "-- name:")

	for _, section := range sections {
		if strings.TrimSpace(section) == "" {
			continue
		}

		lines := strings.Split(section, "\n")
		if len(lines) < 2 {
			continue
		}

		name := strings.TrimSpace(lines[0])
		query := strings.Join(lines[1:], "\n")
		query = strings.TrimSpace(query)
		query = strings.TrimSuffix(query, ";")

		ql.queries[name] = query
	}

	log.Info().Int("count", len(ql.queries)).Msg("Queries loaded successfully")

	return nil
}

func (ql *QueryLoader) Get(name string) (string, bool) {
	query, ok := ql.queries[name]

	return query, ok
}

func (ql *QueryLoader) ExecuteTemplate(name string, data any) (string, []any, error) {
	queryTemplate, ok := ql.Get(name)
	if !ok {
		return "", nil, fmt.Errorf("query %s not found", name)
	}

	tmpl, err := template.New(name).Parse(queryTemplate)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", nil, err
	}

	query := buf.String()

	// Convert named parameters to positional
	return convertNamedToPositional(query, data)
}

func convertNamedToPositional(query string, data any) (string, []any, error) {
	args := make([]any, 0)
	paramMap := make(map[string]any)

	// Extract parameters from data
	if dataMap, ok := data.(map[string]any); ok {
		paramMap = dataMap
	}

	// Replace named parameters with $1, $2, etc.
	paramIndex := 1
	result := query

	for key, value := range paramMap {
		placeholder := "$" + key
		if strings.Contains(result, placeholder) {
			positional := fmt.Sprintf("$%d", paramIndex)
			result = strings.ReplaceAll(result, placeholder, positional)
			args = append(args, value)
			paramIndex++
		}
	}

	return result, args, nil
}
