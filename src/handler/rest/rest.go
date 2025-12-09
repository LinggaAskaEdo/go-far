package rest

import (
	"go-far/src/config"
	"go-far/src/service"
	"sync"

	"github.com/gin-gonic/gin"
)

var onceRestHandler = &sync.Once{}

type rest struct {
	gin  *gin.Engine
	auth config.Auth
	mw   config.Middleware
	svc  *service.Service
}

func InitRestHandler(gin *gin.Engine, auth config.Auth, mw config.Middleware, svc *service.Service) {
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
	// User
	e.gin.POST("/user", e.CreateUser)
	e.gin.GET("/users/:id", e.GetUser)
	e.gin.GET("/users", e.ListUsers)
	e.gin.PUT("/users/:id", e.UpdateUser)
	e.gin.DELETE("/users/:id", e.DeleteUser)
}
