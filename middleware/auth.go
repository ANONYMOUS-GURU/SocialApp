package middleware

import (
	"g_chat/firebase"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func ValidateUserToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var authClient *auth.Client = firebase.GetFirebaseAuthClient()

		authHeader := strings.Split(ctx.Request.Header.Get("Authorization"), " ")

		if len(authHeader) < 2 {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			ctx.Abort()
			return
		}

		token, err := authClient.VerifyIDToken(ctx, authHeader[1])
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized access token"})
			ctx.Abort()
			return
		}
		ctx.Set("userId", token.UID)
		ctx.Next()

		ctx.Set("userId", "token")
		ctx.Next()
	}

}
