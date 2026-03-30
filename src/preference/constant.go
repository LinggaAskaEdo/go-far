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
	CONTEXT_KEY_LOG_TRACE_ID   contextKey = "trace_id"
	CONTEXT_KEY_LOG_SPAN_ID    contextKey = "span_id"
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

	// API Routes
	RouteUsers            string = "/users"
	RouteUsersByID        string = "/users/:id"
	RouteHealth           string = "/health"
	RouteReady            string = "/ready"
	RouteCars             string = "/cars"
	RouteCarsByID         string = "/cars/:id"
	RouteCarsBulk         string = "/cars/bulk"
	RouteCarsOwner        string = "/cars/:id/owner"
	RouteCarsTransfer     string = "/cars/:id/transfer"
	RouteCarsAvailability string = "/cars/availability"
	RouteCarsByUser       string = "/cars/by-user/:user_id"
	RouteCarsByUserCount  string = "/cars/by-user/:user_id/count"

	// Limiter Error Message
	FormatError  string = "Please check the format with your input."
	CommandError string = "The command of first number should > 0"
)
