package query

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/VauntDev/tqla"
	"github.com/rs/zerolog"
)

// QueriesOptions holds query loader configuration
type QueriesOptions struct {
	Path string `yaml:"path"`
}

// QueryLoader loads and manages SQL queries
type QueryLoader struct {
	compileFn func(template string, data any) (string, []any, error)
	queries   map[string]string
}

// InitQueryLoader initializes the query loader
func InitQueryLoader(log zerolog.Logger, opt QueriesOptions) *QueryLoader {
	t, err := tqla.New(
		tqla.WithPlaceHolder(tqla.Dollar),
		tqla.WithFuncMap(template.FuncMap{
			"add": func(x int, y int) int { return x + y },
		}),
	)
	if err != nil {
		log.Panic().Err(err).Msg("Failed to initialize tqla")
	}

	// tqla.Tqla has a Compile method; extract it via reflection since the type is unexported
	v := reflect.ValueOf(t)
	compileMethod := v.MethodByName("Compile")

	ql := &QueryLoader{
		compileFn: func(templateStr string, data any) (string, []any, error) {
			var dataValue reflect.Value
			if data != nil {
				dataValue = reflect.ValueOf(data)
			} else {
				dataValue = reflect.Zero(reflect.TypeOf((*interface{})(nil)).Elem())
			}

			results := compileMethod.Call([]reflect.Value{
				reflect.ValueOf(templateStr),
				dataValue,
			})

			sql := results[0].Interface().(string)
			args := results[1].Interface().([]any)

			var err error
			if !results[2].IsNil() {
				err = results[2].Interface().(error)
			}

			return sql, args, err
		},
		queries: make(map[string]string),
	}

	if err := ql.load(log, opt.Path); err != nil {
		log.Panic().Err(err).Msg("Failed to load queries")
	}

	return ql
}

func (ql *QueryLoader) load(log zerolog.Logger, path string) error {
	files, err := filepath.Glob(filepath.Join(path, "*.sql"))
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no SQL files found in path: %s", path)
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

// Get retrieves a raw query template by name
func (ql *QueryLoader) Get(name string) (string, bool) {
	query, ok := ql.queries[name]
	return query, ok
}

func replaceSortPlaceholders(query string, args []any, _ any) ([]any, string) {
	return args, query
}

// Compile compiles a query template with the provided data, returning the SQL string and arguments
func (ql *QueryLoader) Compile(name string, data any) (string, []any, error) {
	queryTemplate, ok := ql.Get(name)
	if !ok {
		return "", nil, fmt.Errorf("query %s not found", name)
	}

	query, args, err := ql.compileFn(queryTemplate, data)
	if err != nil {
		return "", nil, err
	}

	// Replace sort placeholders with actual values and remove from args
	args, query = replaceSortPlaceholders(query, args, data)

	return query, args, nil
}
