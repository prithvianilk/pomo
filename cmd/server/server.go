package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(DBMiddleware())
	router.GET("/session", GetHandler)
	router.POST("/session", PostHandler)
	router.GET("/maintainance/session", DropAndCreateNew)
	address := fmt.Sprintf(":%s", os.Getenv("PORT"))
	router.Run(address)
}
