package preference

type contextKey string

const (
	// Respnose Status
	STATUS_SUCCESS string = "success"
	STATUS_ERROR   string = "error"

	// Database Type
	MYSQL    string = `mysql`
	POSTGRES string = `postgres`

	// Redis Type
	REDIS_APPS    string = "APPS"
	REDIS_LIMITER string = "LIMITER"
	REDIS_AUTH    string = "AUTH"

	// Logging Context Keys
	CONTEXT_KEY_REQUEST_ID     contextKey = "requestID"
	CONTEXT_KEY_LOG_REQUEST_ID contextKey = "req_id"
	EVENT                      string     = "event"
	METHOD                     string     = "method"
	URL                        string     = "url"
	ADDR                       string     = "addr"
	STATUS                     string     = "status_code"
	LATENCY                    string     = "latency"
	USER_AGENT                 string     = "user_agent"

	// Lang Header
	LANG_EN string = `en`
	LANG_ID string = `id`

	// Custom HTTP Header
	APP_LANG string = `x-app-lang`

	// Cache Control Header
	CacheControl        string = `cache-control`
	CacheMustRevalidate string = `must-revalidate`
)
