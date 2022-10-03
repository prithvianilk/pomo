package main

import (
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(DBMiddleware())
	router.GET("/session", GetHandler)
	router.GET("/session/:name", GetSingleSessionHandler)
	router.POST("/session", PostHandler)
	router.GET("/maintainance/session", DropAndCreateNew)
	address := ":" + os.Getenv("PORT")
	router.Run(address)
}
