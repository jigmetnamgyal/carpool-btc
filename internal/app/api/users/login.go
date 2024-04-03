package users

import (
	"carpool-btc/internal/app/models"
	"carpool-btc/internal/app/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

type Params struct {
	EmailAddress    *string `json:"email_address,omitempty"`
	Password        *string `json:"password,omitempty"`
	ConfirmPassword *string `json:"confirm_password,omitempty"`
	WalletAddress   *string `json:"wallet_address,omitempty"`
}

func Login(c *gin.Context) {
	var userInput Params

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Invalid user data",
			"detail": err.Error(),
		})
		return
	}

	user, token, err := walletAddress(userInput)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to login",
			"details": err.Error(),
		})

		return
	}

	c.Header("Authorization", *token)

	c.JSON(http.StatusOK, user)
	return
}

func walletAddress(ui Params) (*models.User, *string, error) {
	var id int64

	queryString := "SELECT id FROM users u WHERE u.wallet_address = ($1)"
	err := utils.DB.QueryRow(queryString, *ui.WalletAddress).Scan(&id)

	type ResetTokenPayload struct {
		WalletAddress string
	}

	resetTokenPayload := ResetTokenPayload{
		WalletAddress: *ui.WalletAddress,
	}

	token, tokenErr := generateJWTToken(resetTokenPayload, time.Now().Add(time.Hour*24*30).Unix())

	if tokenErr != nil {
		return nil, nil, err
	}

	user, err := GetUserWithWallet(*ui.WalletAddress)

	if err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

func generateJWTToken(payload any, expTime int64) (*string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": payload,
		"exp": expTime,
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}
