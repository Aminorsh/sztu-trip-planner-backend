package middleware

import (
	"fmt"
	"log"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/errors"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		var appErr *errors.AppError

		switch err := recovered.(type) {
		case *errors.AppError:
			appErr = err
		case error:
			appErr = errors.NewInternalServerError(err)
		default:
			appErr = errors.NewInternalServerError(fmt.Errorf("unknown error: %v", recovered))
		}

		if appErr.Internal != nil {
			log.Printf("[ERROR] %s: %v", appErr.Code, appErr.Internal)
		}

		c.JSON(appErr.Status, gin.H{
			"success": false,
			"error": gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
				"details": appErr.Details,
			},
		})
	})
}

func HandleError(c *gin.Context, err error) {
	var appErr *errors.AppError

	switch e := err.(type) {
	case *errors.AppError:
		appErr = e
	default:
		appErr = errors.NewInternalServerError(e)
	}

	c.JSON(appErr.Status, gin.H{
		"success": false,
		"error": gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
			"details": appErr.Details,
		},
	})
}
