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
	RouteAuthRegister     string = "/auth/register"
	RouteAuthLogin        string = "/auth/login"
	RouteAuthRefresh      string = "/auth/refresh"
	RouteUsers            string = "/users"
	RouteUsersV2          string = "/v2/users"
	RouteUsersByID        string = "/users/{id}"
	RouteHealth           string = "/health"
	RouteReady            string = "/ready"
	RouteCars             string = "/cars"
	RouteCarsByID         string = "/cars/{id}"
	RouteCarsBulk         string = "/cars/bulk"
	RouteCarsOwner        string = "/cars/{id}/owner"
	RouteCarsTransfer     string = "/cars/{id}/transfer"
	RouteCarsAvailability string = "/cars/availability"
	RouteCarsByUser       string = "/users/{user_id}/cars"
	RouteCarsByUserCount  string = "/users/{user_id}/cars/count"

	// Limiter Error Message
	FormatError  string = "Please check the format with your input."
	CommandError string = "The command of first number should > 0"

	// Context Keys
	ContextKeyAuthUser contextKey = "auth_user"

	// HTTP Headers
	HeaderContentType               string = "Content-Type"
	HeaderXRateLimitLimitGlobal     string = "X-RateLimit-Limit-global"
	HeaderXRateLimitRemainingGlobal string = "X-RateLimit-Remaining-global"
	HeaderXRateLimitResetGlobal     string = "X-RateLimit-Reset-global"
	HeaderXRateLimitLimitRoute      string = "X-RateLimit-Limit-route"
	HeaderXRateLimitRemainingRoute  string = "X-RateLimit-Remaining-route"
	HeaderXRateLimitResetRoute      string = "X-RateLimit-Reset-route"
	HeaderAuthorization             string = "Authorization"
	HeaderXRequestID                string = "X-Request-ID"
	HeaderXForwardedFor             string = "X-Forwarded-For"
	HeaderXRealIP                   string = "X-Real-IP"
	HeaderAccessControlAllowOrigin  string = "Access-Control-Allow-Origin"
	HeaderAccessControlAllowHeaders string = "Access-Control-Allow-Headers"
	HeaderAccessControlAllowMethods string = "Access-Control-Allow-Methods"
	HeaderXFrameOptions             string = "X-Frame-Options"
	HeaderContentSecurityPolicy     string = "Content-Security-Policy"
	HeaderXXSSProtection            string = "X-XSS-Protection"
	HeaderStrictTransportSecurity   string = "Strict-Transport-Security"
	HeaderReferrerPolicy            string = "Referrer-Policy"
	HeaderXContentTypeOptions       string = "X-Content-Type-Options"
	HeaderPermissionsPolicy         string = "Permissions-Policy"

	// Content Types
	ContentTypeJSON string = "application/json"

	// Auth Error Messages
	ErrInvalidToken string = "Invalid token"

	// Token Constants
	TokenSeparator string = "++"
	TokenKeyPrefix string = "access:"

	// Readiness Status Messages
	StatusReady    string = "ready"
	StatusNotReady string = "not ready"

	// Allowed HTTP Methods
	AllowedMethods string = "GET, POST, PUT, DELETE"

	// CORS Security Values
	CSPValue               string = "default-src 'self'; connect-src *; font-src *; script-src-elem * 'unsafe-inline'; img-src * data:; style-src * 'unsafe-inline';"
	PermissionsPolicyValue string = "geolocation=(),midi=(),sync-xhr=(),microphone=(),camera=(),magnetometer=(),gyroscope=(),fullscreen=(self),payment=()"
)
