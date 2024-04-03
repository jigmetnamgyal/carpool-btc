package btc

import (
	"carpool-btc/internal/app/api/btc/helpers"
	"carpool-btc/internal/app/models"
	"carpool-btc/internal/app/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type CreateWithdrawInput struct {
	Amount string `json:"amount" binding:"required"`
}

func Withdraw(c *gin.Context) {
	var input CreateWithdrawInput
	var walletAddress string
	var balance string

	currentUserId := c.MustGet("current_user").(*models.User).ID

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid deposit request body sent",
			"details": err.Error(),
		})

		return
	}

	queryString := `SELECT balance FROM wallets WHERE user_id = ($1)`

	err := utils.DB.QueryRow(queryString, currentUserId).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get user wallet",
			"details": err.Error(),
		})
		return
	}

	if balance < input.Amount {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient Fund",
		})

		return
	}

	paraphase := "jtn is god"
	timeout := 600

	newUUID := uuid.New()

	unlockPayload := helpers.Payload{
		JsonRPC: 1.0,
		ID:      newUUID.String(),
		Method:  "walletpassphrase",
		Params:  []interface{}{paraphase, timeout},
	}

	_, err = helpers.SendRequest(unlockPayload)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to encrypt wallet",
			"details": err.Error(),
		})

		return
	}

	queryString = `SELECT wallet_address FROM users where id = ($1)`

	err = utils.DB.QueryRow(queryString, currentUserId).Scan(&walletAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Invalid scanning wallet address",
			"details": err.Error(),
		})

		return
	}

	newUUID = uuid.New()

	payload := helpers.Payload{
		JsonRPC: 1.0,
		ID:      newUUID.String(),
		Method:  "sendtoaddress",
		Params:  []interface{}{walletAddress, input.Amount},
	}

	request, err := helpers.SendRequest(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to communicate with bitcoin core",
			"details": err.Error(),
		})

		return
	}

	txid := request["result"].(string)

	newUUID = uuid.New()

	getTxPayload := helpers.Payload{
		JsonRPC: 1.0,
		ID:      newUUID.String(),
		Method:  "gettransaction",
		Params:  []interface{}{txid},
	}

	txRequest, err := helpers.SendRequest(getTxPayload)
	fmt.Println(txRequest)
	fmt.Println(txRequest["result"])
	txResult := txRequest["result"].(map[string]interface{})

	txConfirmation := txResult["confirmations"].(float64)

	if txConfirmation == 0 {
		queryString = `INSERT INTO un_confirmed_balances (user_id, un_confirmed_balance, tx_id, type) VALUES ($1, $2, $3, $4)`

		_, err = utils.DB.Exec(queryString, currentUserId, input.Amount, txid, "withdraw")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create un confirmed record",
				"details": err.Error(),
			})

			return
		}
	}

	queryString = `UPDATE wallets SET balance = balance - ($1) WHERE user_id = ($2)`

	_, err = utils.DB.Exec(queryString, input.Amount, currentUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to deduct amount",
			"details": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully withdrawn the amount"})
}
