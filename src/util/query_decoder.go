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
	if !isSafeForColumn(typ, rawVal) {
		return
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(rawVal)
	case reflect.Int, reflect.Int64:
		parseInt(field, rawVal)
	case reflect.Float64:
		parseFloat(field, rawVal)
	case reflect.Bool:
		parseBool(field, rawVal)
	case reflect.Slice:
		parseStringSlice(field, typ, rawVal)
	}

	if isTimeType(typ) {
		parseTime(field, rawVal)
	}
}

func isSafeForColumn(typ reflect.Type, val string) bool {
	name := strings.ToLower(typ.Name())
	if IsColumnField(name) {
		return IsValidIdentifier(val)
	}

	return true
}

func parseInt(field reflect.Value, val string) {
	if n, err := strconv.ParseInt(val, 10, 64); err == nil {
		field.SetInt(n)
	}
}

func parseFloat(field reflect.Value, val string) {
	if n, err := strconv.ParseFloat(val, 64); err == nil {
		field.SetFloat(n)
	}
}

func parseBool(field reflect.Value, val string) {
	if b, err := strconv.ParseBool(val); err == nil {
		field.SetBool(b)
	}
}

func parseStringSlice(field reflect.Value, typ reflect.Type, val string) {
	if typ.Elem().Kind() != reflect.String {
		return
	}
	parts := strings.Split(val, ",")
	validated := make([]string, 0, len(parts))
	for _, p := range parts {
		if IsValidIdentifier(p) {
			validated = append(validated, p)
		}
	}

	field.Set(reflect.ValueOf(validated))
}

func parseTime(field reflect.Value, val string) {
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		field.Set(reflect.ValueOf(t))
	}
}

func isTimeType(typ reflect.Type) bool {
	return typ == reflect.TypeOf(time.Time{})
}
