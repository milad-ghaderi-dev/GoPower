package handlers

import (
	"GoPower/models"
	"GoPower/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CreateStation(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var station models.PowerStation

		if err := c.ShouldBindJSON(&station); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if err := repo.Create(&station); err != nil {
			if err == models.ErrDuplicateStationID {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "station created",
			"station": station,
		})
	}
}

func GetStationByID(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		station, err := repo.GetByID(id)
		if err != nil {
			if err == models.ErrStationNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return 
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"station": *station})
	}
}

func ListStations(repo repository.StationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		qType := c.Query("type")

		stations := make([]models.PowerStation, 0)
		var err error

		if qType != "" {
			stations, err = repo.GetByType(models.StationType(strings.ToUpper(qType)))
		} else {
			stations, err = repo.ListAll()
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"stations": stations})
	}
}
