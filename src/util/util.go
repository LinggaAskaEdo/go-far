package util

const maxLimit, defaultLimit int64 = 1e4, 10

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
		return 0
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
