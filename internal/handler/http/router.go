package rest

import (
	"net/http"
	"sync"

	"go-far/internal/infra/middleware"
	"go-far/internal/infra/token"
	"go-far/internal/preference"
	"go-far/internal/service"
	"go-far/internal/service/user"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type rest struct {
	mux   *http.ServeMux
	auth  token.Token
	mw    middleware.Middleware
	svc   *service.Service
	usvc  user.UserServiceItf
	sql0  *pgxpool.Pool
	redis *redis.Client
}

var onceRestHandler = &sync.Once{}

func InitHttpHandler(mux *http.ServeMux, auth token.Token, mw middleware.Middleware, svc *service.Service, usvc user.UserServiceItf, sql0 *pgxpool.Pool, redisClient *redis.Client) {
	var e *rest

	onceRestHandler.Do(func() {
		e = &rest{
			mux:   mux,
			auth:  auth,
			mw:    mw,
			svc:   svc,
			usvc:  usvc,
			sql0:  sql0,
			redis: redisClient,
		}

		e.Serve()
	})
}

func (e *rest) Serve() {
	// Health check endpoints (public)
	e.mux.HandleFunc("GET "+preference.RouteHealth, e.Health)
	e.mux.HandleFunc("GET "+preference.RouteReady, e.Ready)

	// Auth routes (public, but rate-limited by IP to prevent brute force)
	authLimiter := e.mw.AuthLimiter()
	e.mux.Handle("POST "+preference.RouteAuthRegister, authLimiter(http.HandlerFunc(e.Register)))
	e.mux.Handle("POST "+preference.RouteAuthLogin, authLimiter(http.HandlerFunc(e.Login)))
	e.mux.Handle("POST "+preference.RouteAuthRefresh, authLimiter(http.HandlerFunc(e.RefreshToken)))

	// Car routes (authenticated, rate-limited by role)
	limiter := e.mw.RoleLimiter()
	e.mux.Handle("POST "+preference.RouteCars, limiter(http.HandlerFunc(e.CreateCar)))
	e.mux.Handle("POST "+preference.RouteCarsBulk, limiter(http.HandlerFunc(e.CreateBulkCars)))
	e.mux.Handle("GET "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.GetCar)))
	e.mux.Handle("GET "+preference.RouteCarsOwner, limiter(http.HandlerFunc(e.GetCarWithOwner)))
	e.mux.Handle("PUT "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.UpdateCar)))
	e.mux.Handle("DELETE "+preference.RouteCarsByID, limiter(http.HandlerFunc(e.DeleteCar)))
	e.mux.Handle("POST "+preference.RouteCarsTransfer, limiter(http.HandlerFunc(e.TransferCarOwnership)))
	e.mux.Handle("PUT "+preference.RouteCarsAvailability, limiter(http.HandlerFunc(e.BulkUpdateAvailability)))

	// User car routes (authenticated, rate-limited by role)
	e.mux.Handle("GET "+preference.RouteCarsByUser, limiter(http.HandlerFunc(e.ListCarsByUser)))
	e.mux.Handle("GET "+preference.RouteCarsByUserCount, limiter(http.HandlerFunc(e.CountCarsByUser)))

	// User routes (authenticated, rate-limited by role)
	e.mux.Handle("POST "+preference.RouteUsers, limiter(http.HandlerFunc(e.CreateUser)))
	e.mux.Handle("GET "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.GetUser)))
	e.mux.Handle("GET "+preference.RouteUsers, limiter(http.HandlerFunc(e.ListUsers)))
	e.mux.Handle("GET "+preference.RouteUsersV2, limiter(http.HandlerFunc(e.ListUsersV2)))
	e.mux.Handle("PUT "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.UpdateUser)))
	e.mux.Handle("DELETE "+preference.RouteUsersByID, limiter(http.HandlerFunc(e.DeleteUser)))
}
