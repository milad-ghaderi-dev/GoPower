package handlers

import (
	"GoPower/models"
	"GoPower/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateConsumer(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var consumer models.ConsumerMeter

		if err := c.ShouldBindJSON(&consumer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		if err := repo.Create(&consumer); err != nil {
			if err == models.ErrDuplicateConsumerID {
				c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":  "consumer created",
			"consumer": consumer,
		})
	}
}

func GetConsumerByID(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
			return
		}
		consumer, err := repo.GetByID(id)
		if err != nil {
			if err == models.ErrConsumerNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return 
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"consumer": *consumer})
	}
}

func ListConsumers(repo repository.ConsumerRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		consumers, err := repo.ListAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"consumers": consumers})
	}
}
