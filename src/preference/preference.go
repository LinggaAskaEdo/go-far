package preference

type contextKey string

const (
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
)
