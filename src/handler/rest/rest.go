package rest

import (
	"sync"

	"go-far/src/config/auth"
	"go-far/src/config/middleware"
	"go-far/src/preference"
	"go-far/src/service"

	"github.com/gin-gonic/gin"
)

var onceRestHandler = &sync.Once{}

type rest struct {
	gin  *gin.Engine
	auth auth.Auth
	mw   middleware.Middleware
	svc  *service.Service
}

func InitRestHandler(gin *gin.Engine, auth auth.Auth, mw middleware.Middleware, svc *service.Service) {
	var e *rest

	onceRestHandler.Do(func() {
		e = &rest{
			gin:  gin,
			auth: auth,
			mw:   mw,
			svc:  svc,
		}

		e.Serve()
	})
}

func (e *rest) Serve() {
	// Health check endpoints
	e.gin.GET(preference.RouteHealth, e.Health)
	e.gin.GET(preference.RouteReady, e.Ready)

	// Car routes
	e.gin.POST(preference.RouteCars, e.CreateCar)
	e.gin.POST(preference.RouteCarsBulk, e.CreateBulkCars)
	e.gin.GET(preference.RouteCarsByID, e.GetCar)
	e.gin.GET(preference.RouteCarsOwner, e.GetCarWithOwner)
	e.gin.PUT(preference.RouteCarsByID, e.UpdateCar)
	e.gin.DELETE(preference.RouteCarsByID, e.DeleteCar)
	e.gin.POST(preference.RouteCarsTransfer, e.TransferCarOwnership)
	e.gin.PUT(preference.RouteCarsAvailability, e.BulkUpdateAvailability)

	// User car routes (using /cars/by-user/:user_id to avoid wildcard conflicts)
	e.gin.GET(preference.RouteCarsByUser, e.ListCarsByUser)
	e.gin.GET(preference.RouteCarsByUserCount, e.CountCarsByUser)

	// User routes
	e.gin.POST(preference.RouteUsers, e.CreateUser)
	e.gin.GET(preference.RouteUsersByID, e.mw.Limiter("1-M", 3), e.GetUser)
	e.gin.GET(preference.RouteUsers, e.ListUsers)
	e.gin.PUT(preference.RouteUsersByID, e.UpdateUser)
	e.gin.DELETE(preference.RouteUsersByID, e.DeleteUser)
}
