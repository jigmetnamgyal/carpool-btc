package middlewares

import (
	"carpool-btc/internal/app/api/users"
	"carpool-btc/internal/app/models"
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RequireAuth(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Authorization Token is not provided",
		})

		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	jwtClaims, errParse := users.ParseJwtToken(tokenString)

	if errParse != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to parse Authorization token",
			"details": errParse.Error(),
		})

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	walletAddress, waOk := jwtClaims["sub"].(map[string]interface{})["WalletAddress"].(string)

	var currentUser *models.User
	var authErr error

	if waOk {
		currentUser, authErr = users.GetUserWithWallet(walletAddress)
	}

	if authErr != nil {
		if errors.Is(authErr, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "user with not found",
				"details": authErr.Error(),
			})

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Failed to find user with email",
			"details": authErr.Error(),
		})

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Set("current_user", currentUser)
	c.Next()
}
