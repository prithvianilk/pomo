package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func getDB() *sqlx.DB {
	db, err := sqlx.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	return db
}

func main() {
	router := gin.Default()
	app := App{db: getDB()}
	router.GET("/session", app.GetHandler)
	router.GET("/session/:name", app.GetSingleSessionHandler)
	router.POST("/session", app.PostHandler)
	router.GET("/name", app.GetSessionNamesHandler)
	router.GET("/maintainance/session", app.DropAndCreateNew)
	address := ":" + os.Getenv("PORT")
	router.Run(address)
}
