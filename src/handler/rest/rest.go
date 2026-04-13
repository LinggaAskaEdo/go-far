package rest

import (
	"net/http"
	"sync"

	"go-far/src/config/middleware"
	"go-far/src/config/token"
	"go-far/src/preference"
	"go-far/src/service"
	"go-far/src/service/user"
)

var onceRestHandler = &sync.Once{}

type rest struct {
	mux  *http.ServeMux
	auth token.Token
	mw   middleware.Middleware
	svc  *service.Service
	usvc user.UserServiceItf
}

func InitRestHandler(mux *http.ServeMux, auth token.Token, mw middleware.Middleware, svc *service.Service, usvc user.UserServiceItf) {
	var e *rest

	onceRestHandler.Do(func() {
		e = &rest{
			mux:  mux,
			auth: auth,
			mw:   mw,
			svc:  svc,
			usvc: usvc,
		}

		e.Serve()
	})
}

func (e *rest) Serve() {
	// Health check endpoints
	e.mux.HandleFunc("GET "+preference.RouteHealth, e.Health)
	e.mux.HandleFunc("GET "+preference.RouteReady, e.Ready)

	// Auth routes
	e.mux.HandleFunc("POST "+preference.RouteAuthRegister, e.Register)
	e.mux.HandleFunc("POST "+preference.RouteAuthLogin, e.Login)
	e.mux.HandleFunc("POST "+preference.RouteAuthRefresh, e.RefreshToken)

	// Car routes
	e.mux.HandleFunc("POST "+preference.RouteCars, e.CreateCar)
	e.mux.HandleFunc("POST "+preference.RouteCarsBulk, e.CreateBulkCars)
	e.mux.HandleFunc("GET "+preference.RouteCarsByID, e.GetCar)
	e.mux.HandleFunc("GET "+preference.RouteCarsOwner, e.GetCarWithOwner)
	e.mux.HandleFunc("PUT "+preference.RouteCarsByID, e.UpdateCar)
	e.mux.HandleFunc("DELETE "+preference.RouteCarsByID, e.DeleteCar)
	e.mux.HandleFunc("POST "+preference.RouteCarsTransfer, e.TransferCarOwnership)
	e.mux.HandleFunc("PUT "+preference.RouteCarsAvailability, e.BulkUpdateAvailability)

	// User car routes
	e.mux.HandleFunc("GET "+preference.RouteCarsByUser, e.ListCarsByUser)
	e.mux.HandleFunc("GET "+preference.RouteCarsByUserCount, e.CountCarsByUser)

	// User routes
	e.mux.HandleFunc("POST "+preference.RouteUsers, e.CreateUser)
	e.mux.HandleFunc("GET "+preference.RouteUsersByID, e.mw.RoleLimiter()(http.HandlerFunc(e.GetUser)).ServeHTTP)
	e.mux.HandleFunc("GET "+preference.RouteUsers, e.ListUsers)
	e.mux.HandleFunc("PUT "+preference.RouteUsersByID, e.UpdateUser)
	e.mux.HandleFunc("DELETE "+preference.RouteUsersByID, e.DeleteUser)
}
