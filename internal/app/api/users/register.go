package users

import (
	"carpool-btc/internal/app/models"
	"carpool-btc/internal/app/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserParams struct {
	EmailAddress  string `json:"email_address" binding:"required"`
	UserName      string `json:"user_name" binding:"required"`
	WalletAddress string `json:"wallet_address" binding:"required"`
	PhoneNumber   string `json:"phone_number"`
}

func Register(c *gin.Context) {
	var userDetails models.User
	var userInput UserParams

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user request body sent",
			"details": err.Error(),
		})

		return
	}

	tx, err := utils.DB.Begin()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to begin db transaction",
			"details": err.Error(),
		})

		return
	}

	queryString := `
		INSERT INTO users 
		    (email_address, user_name, wallet_address, phone_number, role_id)
		VALUES ($1, $2, $3, $4, 1)
		RETURNING id, email_address, user_name, wallet_address, phone_number
	`

	err = tx.QueryRow(
		queryString,
		userInput.EmailAddress, userInput.UserName, userInput.WalletAddress, userInput.PhoneNumber,
	).Scan(
		&userDetails.ID,
		&userDetails.EmailAddress,
		&userDetails.UserName,
		&userDetails.WalletAddress,
		&userDetails.PhoneNumber,
	)

	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to rollback db",
				"details": txErr.Error(),
			})

			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Failed to register user",
			"detail": err.Error(),
		},
		)

		return
	}

	err = tx.Commit()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to commit to db",
			"details": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "user successfully registered",
	})
}
