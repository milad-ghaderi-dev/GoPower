package server

import (
	"GoPower/handlers"
	"GoPower/repository"
	"GoPower/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetupRouter(srv *service.GridService, stRepo repository.StationRepository, csRepo repository.ConsumerRepository) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "OK"}) })
	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.POST("/generation", handlers.RecordGeneration(srv))
			v1.POST("/dispatch", handlers.DispatchEnergy(srv))
			v1.GET("/grid/summary", handlers.GetGridSummary(srv))
			stations := v1.Group("/stations")
			{
				stations.POST("", handlers.CreateStation(stRepo))
				stations.GET("", handlers.ListStations(stRepo))
				stations.GET("/:id", handlers.GetStationByID(stRepo))
			}
			consumers := v1.Group("/consumers")
			{
				consumers.POST("", handlers.CreateConsumer(csRepo))
				consumers.GET("", handlers.ListConsumers(csRepo))
				consumers.GET("/:id", handlers.GetConsumerByID(csRepo))
			}
		}
	}

	return r
}
