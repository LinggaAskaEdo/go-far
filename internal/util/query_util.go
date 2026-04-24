package util

import (
	"regexp"
	"strings"
)

const maxLimit, defaultLimit int64 = 1e4, 10

var (
	reDoubleNewline   = regexp.MustCompile(`\n\s*\n`)
	reNewlineSpaces   = regexp.MustCompile(`\n\s*`)
	reWhitespace      = regexp.MustCompile(`\s+`)
	injectionPattern  = regexp.MustCompile(`(?i)(;|--|\/\*|\*\/|xp_|sp_executesql|exec\s*\(|execute\s*\(|union\s+select)`)
	identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

const MaxIdentifierLength = 64

func ValidatePage(page int64) int64 {
	if page < 1 {
		return 1
	}

	return page
}

func ValidateLimit(limit int64) int64 {
	if limit < 1 {
		return defaultLimit
	} else if limit > maxLimit {
		return maxLimit
	}

	return limit
}

func CleanQuery(query string) string {
	// Replace double newlines with single newline
	query = reDoubleNewline.ReplaceAllString(query, "\n")

	// Replace newline + spaces with a single space
	query = reNewlineSpaces.ReplaceAllString(query, " ")

	// Collapse any remaining whitespace sequences into a single space
	query = reWhitespace.ReplaceAllString(query, " ")

	// Trim leading/trailing spaces
	return strings.TrimSpace(query)
}

func IsValidIdentifier(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > MaxIdentifierLength {
		return false
	}

	if injectionPattern.MatchString(s) {
		return false
	}

	return identifierPattern.MatchString(s)
}

func IsColumnField(name string) bool {
	name = strings.ToLower(name)
	return name == "sortby" || name == "sortdir" || strings.HasSuffix(name, "column") || strings.HasSuffix(name, "field")
}
