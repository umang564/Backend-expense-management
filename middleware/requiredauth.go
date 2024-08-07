package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"app.com/db"
	"app.com/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequiredAuth(c *gin.Context) {
	fmt.Println("this is middleware")

	// Retrieve token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Authorization header not found"})
		c.Abort()
		return
	}

	// Remove "Bearer " prefix from the header value
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Bearer token not found in Authorization header"})
		c.Abort()
		return
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("umang"), nil
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error parsing token: " + err.Error()})
		c.Abort()
		return
	}

	// Check token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid token claims or token is not valid"})
		c.Abort()
		return
	}

	// Check token expiration
	if exp, ok := claims["exp"].(float64); !ok || float64(time.Now().Unix()) > exp {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Token has expired"})
		c.Abort()
		return
	}

	// Lookup user by ID
	var user models.User
	if err := db.Db.First(&user, claims["sub"]).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Error fetching user from database: " + err.Error()})
		c.Abort()
		return
	}

	// Check if user was found
	if user.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID in token"})
		c.Abort()
		return
	}

	// Set user in context
	c.Set("user", user)
	c.Next()
}
