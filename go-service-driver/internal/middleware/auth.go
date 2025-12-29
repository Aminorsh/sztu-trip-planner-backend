package middleware

import (
	"strings"

	"github.com/Aminorsh/sztu-trip-planner-backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "missing or malformed token"})
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := utils.ParseToken(token)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token: " + err.Error()})
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Next()
	}
}
