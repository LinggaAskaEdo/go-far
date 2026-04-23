package middleware

import (
	"net/http"
	"os"
	"slices"
	"strings"

	"go-far/internal/preference"
)

// CORS returns the CORS middleware handler
func (mw *middleware) CORS() func(http.Handler) http.Handler {
	allowedOrigins := getAllowedOrigins()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set(preference.HeaderAccessControlAllowOrigin, origin)
			} else {
				w.Header().Set(preference.HeaderAccessControlAllowOrigin, "*")
			}

			w.Header().Set(preference.HeaderAccessControlAllowHeaders, "*")
			w.Header().Set(preference.HeaderAccessControlAllowMethods, preference.AllowedMethods)
			w.Header().Set(preference.HeaderXFrameOptions, "DENY")
			w.Header().Set(preference.HeaderContentSecurityPolicy, preference.CSPValue)
			w.Header().Set(preference.HeaderXXSSProtection, "1; mode=block")
			w.Header().Set(preference.HeaderStrictTransportSecurity, "max-age=31536000; includeSubDomains; preload")
			w.Header().Set(preference.HeaderReferrerPolicy, "strict-origin")
			w.Header().Set(preference.HeaderXContentTypeOptions, "nosniff")
			w.Header().Set(preference.HeaderPermissionsPolicy, preference.PermissionsPolicyValue)

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if !slices.Contains([]string{"GET", "POST", "PUT", "DELETE"}, r.Method) {
				mw.writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getAllowedOrigins() []string {
	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	if allowedOriginsEnv == "" {
		return []string{"http://localhost:3000", "http://localhost:8181"}
	}

	origins := strings.Split(allowedOriginsEnv, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}

	return origins
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	return slices.Contains(allowedOrigins, origin)
}
