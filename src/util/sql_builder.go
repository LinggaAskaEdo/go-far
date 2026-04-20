package util

import (
	"bytes"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const maxLimit, defaultLimit int64 = 1e4, 10

const (
	one = iota
	many
)

const (
	unknown = iota
	eq
	neq
	in_
	nin
	like
	nlike
	lte
	lt
	gte
	gt
)

type SQLBuilder struct {
	paramTag    string
	colTag      string
	suffixQuery string
	values      map[string]reflect.Value
	page        int64
	limit       int64
}

func NewSQLBuilder(paramTag, colTag string, suffix string, page, limit int64) *SQLBuilder {
	return &SQLBuilder{
		paramTag:    paramTag,
		colTag:      colTag,
		suffixQuery: suffix,
		values:      make(map[string]reflect.Value),
		page:        page,
		limit:       limit,
	}
}

func (qb *SQLBuilder) AliasPrefix(alias string, ptr any) *SQLBuilder {
	p := reflect.ValueOf(ptr)

	if p.Kind() != reflect.Ptr {
		panic(errors.New("passed any should be a pointer"))
	}

	v := p.Elem()
	qb.values[alias] = v

	return qb
}

func (qb *SQLBuilder) Build() (string, []string, []any, error) {
	args := []any{}
	sortBy := []string{}
	sortByDisplay := []string{}
	mapDBcolsByParam := make(map[string]string)
	buff := bytes.NewBufferString("")
	argIdx := 1

	qb.buildColumnMapping(mapDBcolsByParam)
	buff.WriteString(qb.buildWhereClause())

	for table, v := range qb.values {
		alias := qb.getAlias(table)
		for i := 0; i < v.NumField(); i++ {
			arg, colTag, qType := qb.processField(v, i)
			if arg == nil {
				continue
			}

			isSortBy := qb.isSortByField(colTag)
			if isSortBy {
				sortCols, sortDisp := qb.processSortBy(arg, alias, mapDBcolsByParam)
				sortBy = append(sortBy, sortCols...)
				sortByDisplay = append(sortByDisplay, sortDisp...)
				continue
			}

			isPagination := qb.isPaginationField(colTag)
			if isPagination {
				continue
			}

			qb.appendWhereClause(buff, alias, colTag, qType, argIdx)
			args = append(args, arg)
			argIdx++
		}
	}

	qb.appendOrderBy(buff, sortBy)
	qb.appendLimitOffset(buff, &args, &argIdx)

	buff.WriteString(";")

	return buff.String(), sortByDisplay, args, nil
}

func (qb *SQLBuilder) buildColumnMapping(m map[string]string) {
	for table, v := range qb.values {
		alias := qb.getAlias(table)
		for i := 0; i < v.NumField(); i++ {
			tag := v.Type().Field(i).Tag
			if tag.Get(qb.paramTag) != "-" && tag.Get(qb.paramTag) != "" {
				m[alias+tag.Get(qb.paramTag)] = tag.Get(qb.colTag)
			}
		}
	}
}

func (qb *SQLBuilder) getAlias(table string) string {
	if table == "-" || len(table) < 1 {
		return ""
	}

	return table + "."
}

func (qb *SQLBuilder) buildWhereClause() string {
	buff := bytes.NewBufferString(" WHERE 1=1")
	if len(qb.suffixQuery) > 0 {
		buff.WriteString(" AND " + qb.suffixQuery)
	}

	return buff.String()
}

func (qb *SQLBuilder) processField(v reflect.Value, i int) (any, string, int) {
	tag := v.Type().Field(i).Tag
	colTag := tag.Get(qb.colTag)
	paramTag := tag.Get(qb.paramTag)

	if colTag == "" || colTag == "-" {
		return nil, "", 0
	}

	vFieldItf := v.Field(i).Interface()
	qType := unknown
	arg, skip := qb.extractValue(vFieldItf, paramTag, &qType)
	if skip {
		return nil, "", 0
	}

	return arg, colTag, qType
}

func (qb *SQLBuilder) extractValue(vFieldItf any, paramTag string, qType *int) (any, bool) {
	if qb.isSliceType(vFieldItf) {
		return qb.extractSliceValue(vFieldItf, paramTag, qType)
	}

	if qb.isScalarType(vFieldItf) {
		return qb.extractScalarValue(vFieldItf, paramTag, qType)
	}

	return nil, true
}

func (qb *SQLBuilder) isSliceType(v any) bool {
	switch v.(type) {
	case []int64, []string, []float64, []bool, []time.Time:
		return true
	}

	return false
}

func (qb *SQLBuilder) isScalarType(v any) bool {
	switch v.(type) {
	case int, int64, string, float64, bool, time.Time:
		return true
	}

	return false
}

func (qb *SQLBuilder) extractSliceValue(vFieldItf any, paramTag string, qType *int) (any, bool) {
	*qType = qb.getOperator(many, paramTag)
	switch f := vFieldItf.(type) {
	case []int64:
		if len(f) > 0 {
			return f, false
		}
	case []string:
		if len(f) > 0 {
			return f, false
		}
	case []float64:
		if len(f) > 0 {
			return f, false
		}
	case []bool:
		if len(f) > 0 {
			return f, false
		}
	case []time.Time:
		if len(f) > 0 {
			return f, false
		}
	}

	return nil, true
}

func (qb *SQLBuilder) extractScalarValue(vFieldItf any, paramTag string, qType *int) (any, bool) {
	*qType = qb.getOperator(one, paramTag)
	switch f := vFieldItf.(type) {
	case int:
		if f > 0 {
			return int64(f), false
		}
	case int64:
		if f > 0 {
			return f, false
		}
	case string:
		if len(f) > 0 {
			qb.applyLikeModifier(f, qType)
			return f, false
		}
	case float64:
		if f > 0 {
			return f, false
		}
	case bool:
		if f {
			return f, false
		}
	case time.Time:
		if !f.IsZero() {
			return f, false
		}
	}

	return nil, true
}

func (qb *SQLBuilder) applyLikeModifier(s string, qType *int) {
	if !strings.Contains(s, "%") {
		return
	}
	if *qType == eq {
		*qType = like
	} else {
		*qType = nlike
	}
}

func (qb *SQLBuilder) isSortByField(colTag string) bool {
	switch colTag {
	case "sortby", "orderby", "sort_by", "order_by", "sort-by", "order-by":
		return true
	}

	return false
}

func (qb *SQLBuilder) isPaginationField(colTag string) bool {
	switch colTag {
	case "page", "size", "limit", "offset":
		return true
	}

	return false
}

func (qb *SQLBuilder) processSortBy(arg any, alias string, mapDBcolsByParam map[string]string) ([]string, []string) {
	v, ok := arg.(string)
	if !ok || v == "" {
		return nil, nil
	}

	reg := regexp.MustCompile(`(?P<sign>-)?(?P<col>[a-zA-Z_]+),?`)
	if !reg.MatchString(v) {
		return nil, nil
	}

	return qb.parseSortFields(v, reg, alias, mapDBcolsByParam)
}

func (qb *SQLBuilder) parseSortFields(v string, reg *regexp.Regexp, alias string, mapDBcolsByParam map[string]string) ([]string, []string) {
	sortBy := []string{}
	sortByDisplay := []string{}
	for _, s := range strings.Split(v, ",") {
		match := reg.FindStringSubmatch(s)
		if match == nil {
			continue
		}

		col, sort := qb.extractColumnAndSort(match, reg, alias)
		if col == "" {
			continue
		}

		if colDB, ok := mapDBcolsByParam[alias+col]; ok {
			sortBy = append(sortBy, alias+colDB+" "+sort)
			sortByDisplay = append(sortByDisplay, alias+col+" "+sort)
		}
	}

	return sortBy, sortByDisplay
}

func (qb *SQLBuilder) extractColumnAndSort(match []string, reg *regexp.Regexp, _ string) (string, string) {
	sort := "asc"
	col := ""
	for i, name := range reg.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}

		if match[i] == "-" {
			sort = "desc"
		} else if name == "col" {
			col = match[i]
		}
	}

	return col, sort
}

func (qb *SQLBuilder) appendWhereClause(buff *bytes.Buffer, alias, colTag string, qType int, argIdx int) {
	switch qType {
	case eq:
		buff.WriteString(" AND " + alias + colTag + "=$" + strconv.Itoa(argIdx))
	case neq:
		buff.WriteString(" AND " + alias + colTag + "!=$" + strconv.Itoa(argIdx))
	case gte:
		buff.WriteString(" AND " + alias + colTag + ">=$" + strconv.Itoa(argIdx))
	case gt:
		buff.WriteString(" AND " + alias + colTag + ">$" + strconv.Itoa(argIdx))
	case lte:
		buff.WriteString(" AND " + alias + colTag + "<=$" + strconv.Itoa(argIdx))
	case lt:
		buff.WriteString(" AND " + alias + colTag + "<$" + strconv.Itoa(argIdx))
	case like:
		buff.WriteString(" AND " + alias + colTag + " LIKE $" + strconv.Itoa(argIdx))
	case nlike:
		buff.WriteString(" AND " + alias + colTag + " NOT LIKE $" + strconv.Itoa(argIdx))
	case in_:
		buff.WriteString(" AND " + alias + colTag + " IN ($" + strconv.Itoa(argIdx) + ")")
	case nin:
		buff.WriteString(" AND " + alias + colTag + " NOT IN ($" + strconv.Itoa(argIdx) + ")")
	}
}

func (qb *SQLBuilder) appendOrderBy(buff *bytes.Buffer, sortBy []string) {
	if len(sortBy) > 0 {
		buff.WriteString(" ORDER BY " + strings.Join(sortBy, ", "))
	}
}

func (qb *SQLBuilder) appendLimitOffset(buff *bytes.Buffer, args *[]any, argIdx *int) {
	qb.limit = ValidateLimit(qb.limit)
	qb.page = ValidatePage(qb.page)

	if qb.page > 0 || qb.limit > 0 {
		offset := getOffset(qb.page, qb.limit)
		buff.WriteString(" LIMIT $" + strconv.Itoa(*argIdx))
		*args = append(*args, qb.limit)
		*argIdx++
		buff.WriteString(" OFFSET $" + strconv.Itoa(*argIdx))
		*args = append(*args, offset)
		*argIdx++
	}
}

func getOffset(page, limit int64) int64 {
	return (page - 1) * limit
}

func (qb *SQLBuilder) getOperator(valType int, paramTag string) int {
	if valType == one {
		if strings.Contains(paramTag, "__gte") {
			return gte
		} else if strings.Contains(paramTag, "__lte") {
			return lte
		} else if strings.Contains(paramTag, "__lt") {
			return lt
		} else if strings.Contains(paramTag, "__gt") {
			return gt
		} else if strings.Contains(paramTag, "__neq") {
			return neq
		}

		return eq
	}

	if strings.Contains(paramTag, "__nin") {
		return nin
	}

	return in_
}

func ValidateLimit(limit int64) int64 {
	if limit < 1 {
		return defaultLimit
	} else if limit > maxLimit {
		return maxLimit
	}

	return limit
}

func ValidatePage(page int64) int64 {
	if page < 1 {
		return 1
	}

	return page
}

func ValidateSortBy(sort string) string {
	if sort == "" {
		return "name"
	}

	return sort
}

func ValidateSortDir(sort string) string {
	if sort == "" {
		return "ASC"
	}

	return sort
}
