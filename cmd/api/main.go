package main

import (
	"go-far/internal/app"
)

// @title			Go-Far
// @version		1.14.0
// @description	A production-ready RESTful API built with Go following Domain-Driven Design principles, featuring PostgreSQL, Redis, JWT authentication, role-based rate limiting, and OpenTelemetry tracing.
// @termsOfService	http://swagger.io/terms/
// @contact.name	API Support
// @contact.url	http://www.swagger.io/support
// @contact.email	support@swagger.io
// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//
// @host			localhost:8181
// @schemes		http https
func main() {
	app.Run()
}
