package btc

import (
	"carpool-btc/internal/app/api/btc/helpers"
	"carpool-btc/internal/app/models"
	"carpool-btc/internal/app/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"net/http"
)

type CreateDepositInput struct {
	TxID   string          `json:"tx_id" binding:"required"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
}

// todo: update balace after confirmation only from btc response

func Deposit(c *gin.Context) {
	var input CreateDepositInput

	currentUserId := c.MustGet("current_user").(*models.User).ID

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid deposit request body sent",
			"details": err.Error(),
		})

		return
	}

	newUUID := uuid.New()

	payload := helpers.Payload{
		JsonRPC: 1.0,
		ID:      newUUID.String(),
		Method:  "gettransaction",
		Params:  []interface{}{input.TxID},
	}

	request, err := helpers.SendRequest(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to communicate with bitcoin core",
			"details": err.Error(),
		})

		return
	}

	result := request["result"].(map[string]interface{})
	txConfirmation := result["confirmations"].(float64)

	queryString := `INSERT INTO un_confirmed_balances (user_id, un_confirmed_balance, tx_id) VALUES ($1, $2, $3)`

	_, err = utils.DB.Exec(queryString, currentUserId, input.Amount, input.TxID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create btc record",
			"details": err.Error(),
		})

		return
	}

	confirmedTx := false

	if txConfirmation > 2 {
		var userWalletCount int64

		confirmedTx = true
		queryString = `SELECT COUNT(*) FROM wallets WHERE user_id = ($1)`

		err = utils.DB.QueryRow(queryString, currentUserId).Scan(&userWalletCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get the count of user wallet",
				"details": err.Error(),
			})

			return
		}

		if userWalletCount == 1 {
			queryString = `UPDATE wallets SET balance = balance + ($1) WHERE user_id = ($2)`
			_, updateErr := utils.DB.Exec(queryString, input.Amount, currentUserId)

			if updateErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to update user wallet balance",
					"details": updateErr.Error(),
				})

				return
			}
		} else {
			queryString = `INSERT INTO wallets (user_id, balance) VALUES ($1, $2)`
			_, insertErr := utils.DB.Exec(queryString, currentUserId, input.Amount)

			if insertErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "failed to insert user wallet balance",
					"details": insertErr.Error(),
				})

				return
			}
		}

		queryString = `DELETE FROM un_confirmed_balances WHERE tx_id = ($1)`
		_, deleteErr := utils.DB.Exec(queryString, input.TxID)

		if deleteErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to delete un confirmed balances",
				"details": deleteErr.Error(),
			})

			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_confirmed":   confirmedTx,
		"transaction_details": request,
	})
}
