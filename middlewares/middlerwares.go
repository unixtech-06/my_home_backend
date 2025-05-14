package middlewares

import (
	"backend/utils/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware はJWT認証を行うミドルウェアを返します
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.Valid(c)

		if err != nil {
			// より詳細なエラーメッセージを返す
			errorMessage := "認証に失敗しました"
			if err.Error() != "" {
				errorMessage = err.Error()
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
			c.Abort()
			return
		}

		c.Next()
	}
}
