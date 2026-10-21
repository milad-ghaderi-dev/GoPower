package handlers

import (
	"GoPower/models"
	"GoPower/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecordGenerationRequest struct {
	StationID string  `json:"station_id"`
	OutputMW  float64 `json:"output_mw"`
}

type DispatchRequest struct {
	StationID  string  `json:"station_id"`
	ConsumerID string  `json:"consumer_id"`
	AmountMW   float64 `json:"amount_mw"`
}

func RecordGeneration(srv *service.GridService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rec RecordGenerationRequest
		if err := c.ShouldBindJSON(&rec); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := srv.RecordGeneration(rec.StationID, rec.OutputMW); err != nil {
			if err == models.ErrStationNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "generation recorded"})
	}
}

func DispatchEnergy(srv *service.GridService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DispatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rec, err := srv.DispatchEnergy(req.StationID, req.ConsumerID, req.AmountMW)
		if err != nil {
			if err == models.ErrStationNotFound || err == models.ErrConsumerNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if err == models.ErrInvalidAmount {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "dispatch successful",
			"record":  *rec,
		})
	}
}
