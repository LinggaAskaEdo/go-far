package rest

import (
	exception "go-far/src/errors"

	"github.com/gin-gonic/gin"
)

func (e *rest) httpRespSuccess(c *gin.Context, statusCode int, resp interface{}) {
	c.JSON(statusCode, resp)
}

func (e *rest) httpRespError(c *gin.Context, appErr *exception.AppError) {
	c.JSON(appErr.Status, gin.H{
		"code":    appErr.Code,
		"message": appErr.Message,
		"status":  appErr.Status,
	})
}
