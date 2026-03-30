package query

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

// QueriesOptions holds query loader configuration
type QueriesOptions struct {
	Path string `yaml:"path"`
}

// QueryLoader loads and manages SQL queries
type QueryLoader struct {
	queries  map[string]string
	filePath string
}

// InitQueryLoader initializes the query loader
func InitQueryLoader(log zerolog.Logger, opt QueriesOptions) *QueryLoader {
	ql := &QueryLoader{
		queries:  make(map[string]string),
		filePath: opt.Path,
	}

	if err := ql.load(log); err != nil {
		log.Panic().Err(err).Msg("Failed to load queries")
	}

	return ql
}

func (ql *QueryLoader) load(log zerolog.Logger) error {
	files, err := filepath.Glob(filepath.Join(ql.filePath, "*.sql"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no SQL files found in path: %s", ql.filePath)
	}

	for _, file := range files {
		if err := ql.loadFile(log, file); err != nil {
			return fmt.Errorf("failed to load file %s: %w", file, err)
		}
	}

	log.Debug().Msg("Queries loaded successfully, total queries: " + fmt.Sprint(len(ql.queries)))

	return nil
}

func (ql *QueryLoader) loadFile(log zerolog.Logger, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	sections := strings.SplitSeq(content, "-- name:")

	for section := range sections {
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

	log.Debug().Str("file", filepath.Base(filePath)).Msg("Loaded queries from file")

	return nil
}

// Get retrieves a query by name
func (ql *QueryLoader) Get(name string) (string, bool) {
	query, ok := ql.queries[name]
	return query, ok
}

// ExecuteTemplate executes a query template with the provided data
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
	return convertNamedToPositional(query, data)
}

func convertNamedToPositional(query string, data any) (string, []any, error) {
	args := make([]any, 0)
	paramMap := make(map[string]any)

	if dataMap, ok := data.(map[string]any); ok {
		paramMap = dataMap
	}

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
