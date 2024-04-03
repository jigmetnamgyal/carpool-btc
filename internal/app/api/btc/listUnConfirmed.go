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

func ListUnConfirmed(c *gin.Context) {
	currentUserId := c.MustGet("current_user").(*models.User).ID

	var unConfirmedTxs []models.UnConfirmedTx

	queryString := `SELECT * FROM un_confirmed_balances WHERE user_id = ($1) ORDER BY created_at DESC`

	row, err := utils.DB.Query(queryString, currentUserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to query unconfirmed tx",
			"details": err.Error(),
		})
	}

	for row.Next() {
		var unConfirmedTx models.UnConfirmedTx

		err = row.Scan(
			&unConfirmedTx.ID,
			&unConfirmedTx.UserId,
			&unConfirmedTx.UnConfirmedBalances,
			&unConfirmedTx.CreatedAt,
			&unConfirmedTx.UpdatedAt,
			&unConfirmedTx.TxID,
			&unConfirmedTx.Type,
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error":   "Invalid scanning the un confirmed tx values from db",
				"Details": err.Error(),
			})

			return
		}

		newUUID := uuid.New()

		payload := helpers.Payload{
			JsonRPC: 1.0,
			ID:      newUUID.String(),
			Method:  "gettransaction",
			Params:  []interface{}{unConfirmedTx.TxID},
		}

		request, reqErr := helpers.SendRequest(payload)

		if reqErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to communicate with bitcoin core",
				"details": reqErr.Error(),
			})

			return
		}

		result := request["result"].(map[string]interface{})
		txConfirmation := result["confirmations"].(float64)

		unConfirmedTx.Confirmed = false

		if txConfirmation > 0 {
			var userWalletCount int64

			queryString = `SELECT COUNT(*) FROM wallets WHERE user_id = ($1)`

			err = utils.DB.QueryRow(queryString, currentUserId).Scan(&userWalletCount)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to get the count of user wallet",
					"details": err.Error(),
				})

				return
			}

			var unconfirmedBalance decimal.Decimal

			queryString = `SELECT un_confirmed_balance FROM un_confirmed_balances WHERE tx_id = ($1)`

			err = utils.DB.QueryRow(queryString, unConfirmedTx.TxID).Scan(&unconfirmedBalance)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "failed to query un confirmed balances",
					"details": err.Error(),
				})

				return
			}

			if unConfirmedTx.Type == "deposit" {
				if userWalletCount == 1 {
					queryString = `UPDATE wallets SET balance = balance + ($1) WHERE user_id = ($2)`
					_, updateErr := utils.DB.Exec(queryString, unconfirmedBalance, currentUserId)

					if updateErr != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "Failed to update user wallet balance",
							"details": updateErr.Error(),
						})

						return
					}
				} else {
					queryString = `INSERT INTO wallets (user_id, balance) VALUES ($1, $2)`
					_, insertErr := utils.DB.Exec(queryString, currentUserId, unconfirmedBalance)

					if insertErr != nil {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error":   "failed to insert user wallet balance",
							"details": insertErr.Error(),
						})

						return
					}
				}
			}

			queryString = `DELETE FROM un_confirmed_balances WHERE tx_id = ($1)`
			_, deleteErr := utils.DB.Exec(queryString, unConfirmedTx.TxID)

			if deleteErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to delete un confirmed balances",
					"details": deleteErr.Error(),
				})

				return
			}
			unConfirmedTx.Confirmed = true
		}

		unConfirmedTxs = append(unConfirmedTxs, unConfirmedTx)
	}

	c.JSON(http.StatusOK, gin.H{"unConfirmedTxs": unConfirmedTxs})
}
