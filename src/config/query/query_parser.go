package query

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"text/template"

	x "go-far/src/model/errors"
	"go-far/src/util"
)

var validatedTypes sync.Map // cache for validated struct types

func (ql *QueryLoader) Compile(name string, data any) (query string, args []any, err error) {
	tmpl, ok := ql.templates[name]
	if !ok {
		return "", nil, x.NewWithCode(x.CodeSQLQueryNotFound, fmt.Sprintf("query %s not found", name))
	}

	if err := validateData(data); err != nil {
		return "", nil, err
	}

	return ql.compileTemplate(tmpl, data)
}

func validateData(data any) error {
	if data == nil {
		return nil
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	if _, ok := validatedTypes.Load(t); ok {
		return nil
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		name := strings.ToLower(field.Name)
		if util.IsColumnField(name) {
			val := v.Field(i).String()
			if val != "" && !util.IsValidIdentifier(val) {
				return x.NewWithCode(x.CodeInvalidIdentifier, fmt.Sprintf("invalid identifier: %s", field.Name))
			}
		}
	}

	validatedTypes.Store(t, struct{}{})
	return nil
}

func (ql *QueryLoader) compileTemplate(tmpl *template.Template, data any) (query string, args []any, err error) {
	var sb strings.Builder

	funcMap := template.FuncMap{
		"arg": func(v any) (string, error) {
			args = append(args, v)
			return "$" + strconv.Itoa(len(args)), nil
		},
		"raw": func(v any) (string, error) {
			// Output value directly – no placeholder. Caller must validate!
			return fmt.Sprintf("%v", v), nil
		},
		"eq":  reflect.DeepEqual,
		"ne":  func(a, b any) bool { return !reflect.DeepEqual(a, b) },
		"gt":  func(a, b int) bool { return a > b },
		"lt":  func(a, b int) bool { return a < b },
		"gte": func(a, b int) bool { return a >= b },
		"lte": func(a, b int) bool { return a <= b },
	}

	cloned, err := tmpl.Clone()
	if err != nil {
		return "", nil, err
	}

	cloned = cloned.Funcs(funcMap)

	if err := cloned.Execute(&sb, data); err != nil {
		return "", nil, x.WrapWithCode(err, x.CodeTemplateExecute, fmt.Sprintf("execute template %s", tmpl.Name()))
	}

	return sb.String(), args, nil
}
