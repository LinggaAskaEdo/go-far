package util

import (
	"regexp"
	"strings"
)

var (
	InjectionPattern  = regexp.MustCompile(`(?i)(;|--|\/\*|\*\/|xp_|sp_executesql|exec\s*\(|execute\s*\(|union\s+select)`)
	IdentifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

const MaxIdentifierLength = 64

func IsValidIdentifier(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > MaxIdentifierLength {
		return false
	}

	if InjectionPattern.MatchString(s) {
		return false
	}

	return IdentifierPattern.MatchString(s)
}

func IsColumnField(name string) bool {
	name = strings.ToLower(name)
	return name == "sortby" || name == "sortdir" || strings.HasSuffix(name, "column") || strings.HasSuffix(name, "field")
}

func CleanQuery(query string) string {
	re := regexp.MustCompile(`\n\s*\n`)
	query = re.ReplaceAllString(query, "\n")

	re = regexp.MustCompile(`\n\s*`)
	query = re.ReplaceAllString(query, " ")

	re = regexp.MustCompile(`\s+`)
	query = re.ReplaceAllString(query, " ")

	query = strings.TrimSpace(query)

	return query
}
