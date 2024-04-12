package carpool

import (
	"carpool-btc/internal/app/models"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Create(c *gin.Context) {
	var carPoolInput models.Carpool
	if err := c.ShouldBindJSON(&carPoolInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid carpool input sent in body",
			"details": err.Error(),
		})

		return
	}

	//queryString := `INSERT INTO carpools
	//(departure_point, destination, departure_time, available_seats, price_per_seat, payment_method)
	//VALUES ($1, )
	//`
}
