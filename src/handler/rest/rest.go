package rest

import (
	"net/http"
	"sync"

	"go-far/src/config/middleware"
	"go-far/src/config/token"
	"go-far/src/preference"
	"go-far/src/service"
	"go-far/src/service/user"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

var onceRestHandler = &sync.Once{}

type rest struct {
	mux   *http.ServeMux
	auth  token.Token
	mw    middleware.Middleware
	svc   *service.Service
	usvc  user.UserServiceItf
	sql0  *sqlx.DB
	redis *redis.Client
}

func InitRestHandler(mux *http.ServeMux, auth token.Token, mw middleware.Middleware, svc *service.Service, usvc user.UserServiceItf, sql0 *sqlx.DB, redis *redis.Client) {
	var e *rest

	onceRestHandler.Do(func() {
		e = &rest{
			mux:   mux,
			auth:  auth,
			mw:    mw,
			svc:   svc,
			usvc:  usvc,
			sql0:  sql0,
			redis: redis,
		}

		e.Serve()
	})
}

func (e *rest) Serve() {
	// Health check endpoints (public)
	e.mux.HandleFunc("GET "+preference.RouteHealth, e.Health)
	e.mux.HandleFunc("GET "+preference.RouteReady, e.Ready)

	// Auth routes (public)
	e.mux.HandleFunc("POST "+preference.RouteAuthRegister, e.Register)
	e.mux.HandleFunc("POST "+preference.RouteAuthLogin, e.Login)
	e.mux.HandleFunc("POST "+preference.RouteAuthRefresh, e.RefreshToken)

	// Car routes (authenticated, rate-limited by role)
	limiter := e.mw.RoleLimiter()
	e.mux.HandleFunc("POST "+preference.RouteCars, limiter(http.HandlerFunc(e.CreateCar)).ServeHTTP)
	e.mux.HandleFunc("POST "+preference.RouteCarsBulk, limiter(http.HandlerFunc(e.CreateBulkCars)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.GetCar)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteCarsOwner, limiter(http.HandlerFunc(e.GetCarWithOwner)).ServeHTTP)
	e.mux.HandleFunc("PUT "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.UpdateCar)).ServeHTTP)
	e.mux.HandleFunc("DELETE "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.DeleteCar)).ServeHTTP)
	e.mux.HandleFunc("POST "+preference.RouteCarsTransfer, limiter(http.HandlerFunc(e.TransferCarOwnership)).ServeHTTP)
	e.mux.HandleFunc("PUT "+preference.RouteCarsAvailability, limiter(http.HandlerFunc(e.BulkUpdateAvailability)).ServeHTTP)

	// User car routes (authenticated, rate-limited by role)
	e.mux.HandleFunc("GET "+preference.RouteCarsByUser, limiter(http.HandlerFunc(e.ListCarsByUser)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteCarsByUserCount, limiter(http.HandlerFunc(e.CountCarsByUser)).ServeHTTP)

	// User routes (authenticated, rate-limited by role)
	e.mux.HandleFunc("POST "+preference.RouteUsers, limiter(http.HandlerFunc(e.CreateUser)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.GetUser)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteUsers, limiter(http.HandlerFunc(e.ListUsers)).ServeHTTP)
	e.mux.HandleFunc("PUT "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.UpdateUser)).ServeHTTP)
	e.mux.HandleFunc("DELETE "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.DeleteUser)).ServeHTTP)
}
