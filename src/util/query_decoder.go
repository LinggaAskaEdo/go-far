package util

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func DecodeQuery[T any](q url.Values) T {
	var result T
	typ := reflect.TypeOf(result)
	val := reflect.ValueOf(&result).Elem()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag

		param := tag.Get("param")
		if param == "" {
			param = tag.Get("form")
		}

		if param == "" {
			param = strings.ToLower(field.Name)
		}

		if param == "-" {
			continue
		}

		if !val.Field(i).CanSet() {
			continue
		}

		rawVal := q.Get(param)
		if rawVal == "" {
			continue
		}

		setField(val.Field(i), rawVal, field.Type)
	}

	return result
}

func setField(field reflect.Value, rawVal string, typ reflect.Type) {
	switch typ.Kind() {
	case reflect.String:
		field.SetString(rawVal)
	case reflect.Int, reflect.Int64:
		if n, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
			field.SetInt(n)
		}
	case reflect.Float64:
		if n, err := strconv.ParseFloat(rawVal, 64); err == nil {
			field.SetFloat(n)
		}
	case reflect.Bool:
		if b, err := strconv.ParseBool(rawVal); err == nil {
			field.SetBool(b)
		}
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.String {
			parts := strings.Split(rawVal, ",")
			field.Set(reflect.ValueOf(parts))
		}
	}

	if typ == reflect.TypeOf(time.Time{}) {
		if t, err := time.Parse(time.RFC3339, rawVal); err == nil {
			field.Set(reflect.ValueOf(t))
		}
	}
}
