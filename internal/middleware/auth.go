package middleware

import (
	"event-tracking-service/pkg/common"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func RequiredToken(secretKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractTokenFromHeaderString(c)
		if err != nil {
			common.WriteErrorUnauthorizedResponse(c, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		isValid, err := validateToken(token, secretKey)
		if err != nil {
			common.WriteErrorUnauthorizedResponse(c, "Unauthorized", err.Error())
			c.Abort()
			return
		}

		if !isValid {
			common.WriteErrorUnauthorizedResponse(c, "Unauthorized", "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}

func extractTokenFromHeaderString(c *gin.Context) (string, error) {
	s := c.GetHeader("Authorization")
	parts := strings.Split(s, " ")

	if parts[0] != "Bearer" || len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return "", errors.New("token invalid")
	}

	return parts[1], nil
}

func validateToken(tokenString string, secretKey []byte) (bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return false, err
	}

	return token.Valid, nil
}
