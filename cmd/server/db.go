package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func DBMiddleware() gin.HandlerFunc {
	db := getDB()
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func ParseDB(c *gin.Context) *sqlx.DB {
	db, _ := c.MustGet("db").(*sqlx.DB)
	return db
}

func getDB() *sqlx.DB {
	db, err := sqlx.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		panic(err)
	}
	return db
}
