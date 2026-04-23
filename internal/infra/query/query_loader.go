package query

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	x "go-far/internal/model/errors"
)

// func (ql *QueryLoader) load(path string) error {
// 	files, err := filepath.Glob(filepath.Join(path, "*.sql"))
// 	if err != nil {
// 		return err
// 	}

// 	if len(files) == 0 {
// 		return x.NewWithCode(x.CodeFileNotFound, fmt.Sprintf("no SQL files found in path: %s", path))
// 	}

// 	for _, file := range files {
// 		if err := ql.loadFile(file); err != nil {
// 			return x.WrapWithCode(err, x.CodeFileRead, fmt.Sprintf("failed to load file %s", file))
// 		}
// 	}

// 	return nil
// }

func (ql *QueryLoader) load(baseDir string) error {
	// Open a root that restricts all operations to baseDir
	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return fmt.Errorf("failed to open root %s: %w", baseDir, err)
	}
	defer func() {
		_ = root.Close()
	}()

	// Walk the directory using root.ReadDir (or filepath.WalkDir with root paths)
	// For each file, call loadFile with the relative path
	return ql.walkAndLoad(root, ".")
	// return err
}

func (ql *QueryLoader) walkAndLoad(root *os.Root, relPath string) error {
	// Open the directory using root (prevents path traversal)
	dir, err := root.Open(relPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = dir.Close()
	}()

	// Read all directory entries
	entries, err := dir.ReadDir(-1)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullRel := filepath.Join(relPath, entry.Name())
		if entry.IsDir() {
			if err := ql.walkAndLoad(root, fullRel); err != nil {
				return err
			}
			continue
		}
		// Load file content
		if err := ql.loadFileFromRoot(root, fullRel); err != nil {
			return err
		}
	}

	return nil
}

func (ql *QueryLoader) loadFileFromRoot(root *os.Root, relPath string) error {
	// Open the file via root (safe from traversal)
	f, err := root.Open(relPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	content := string(data)
	sections := strings.SplitSeq(content, "-- name:")

	for section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		lines := strings.SplitN(section, "\n", 2)
		if len(lines) < 2 {
			continue
		}

		name := strings.TrimSpace(lines[0])
		sqlText := strings.TrimSpace(lines[1])
		sqlText = strings.TrimSuffix(sqlText, ";")

		// Check for duplicate query name across files
		if existingFile, ok := ql.fileMap[name]; ok {
			return x.NewWithCode(x.CodeDuplicateQuery, fmt.Sprintf("duplicate query name %q found in %s (already defined in %s)", name, relPath, existingFile))
		}

		tmpl, err := template.New(name).Funcs(ql.baseFuncMap()).Option("missingkey=error").Parse(sqlText)
		if err != nil {
			return x.WrapWithCode(err, x.CodeTemplateParse, fmt.Sprintf("parse template %s", name))
		}

		ql.templates[name] = tmpl
		ql.rawSQL[name] = sqlText
		ql.fileMap[name] = relPath
		ql.log.Debug().Str("file", filepath.Base(relPath)).Str("query", name).Msg("Loaded query")
	}

	return nil
}

// func (ql *QueryLoader) loadFile(filePath string) error {
// 	data, err := os.ReadFile(filePath)
// 	if err != nil {
// 		return err
// 	}

// 	content := string(data)
// 	sections := strings.SplitSeq(content, "-- name:")

// 	for section := range sections {
// 		section = strings.TrimSpace(section)
// 		if section == "" {
// 			continue
// 		}

// 		lines := strings.SplitN(section, "\n", 2)
// 		if len(lines) < 2 {
// 			continue
// 		}

// 		name := strings.TrimSpace(lines[0])
// 		sqlText := strings.TrimSpace(lines[1])
// 		sqlText = strings.TrimSuffix(sqlText, ";")

// 		// Check for duplicate query name across files
// 		if existingFile, ok := ql.fileMap[name]; ok {
// 			return x.NewWithCode(x.CodeDuplicateQuery, fmt.Sprintf("duplicate query name %q found in %s (already defined in %s)", name, filePath, existingFile))
// 		}

// 		tmpl, err := template.New(name).Funcs(ql.baseFuncMap()).Option("missingkey=error").Parse(sqlText)
// 		if err != nil {
// 			return x.WrapWithCode(err, x.CodeTemplateParse, fmt.Sprintf("parse template %s", name))
// 		}

// 		ql.templates[name] = tmpl
// 		ql.rawSQL[name] = sqlText
// 		ql.fileMap[name] = filePath
// 		ql.log.Debug().Str("file", filepath.Base(filePath)).Str("query", name).Msg("Loaded query")
// 	}

// 	return nil
// }

func (ql *QueryLoader) baseFuncMap() template.FuncMap {
	return template.FuncMap{
		"eq":  reflect.DeepEqual,
		"ne":  func(a, b any) bool { return !reflect.DeepEqual(a, b) },
		"gt":  func(a, b int) bool { return a > b },
		"lt":  func(a, b int) bool { return a < b },
		"gte": func(a, b int) bool { return a >= b },
		"lte": func(a, b int) bool { return a <= b },
		"arg": func(v any) (string, error) {
			// Dummy placeholder; actual argument collection happens at execution.
			return "$1", nil
		},
		"raw": func(v any) (string, error) {
			return fmt.Sprintf("%v", v), nil
		},
	}
}
