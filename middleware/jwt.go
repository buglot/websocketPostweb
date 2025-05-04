package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		hmacSampleSecret := []byte(os.Getenv("JWT_SECRAT_KEY"))
		tokenString := strings.Replace(header, "Bearer ", "", 1)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return hmacSampleSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx.Set("userID", claims["userID"])
		} else {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": err.Error()})
		}
		ctx.Next()
	}

}
