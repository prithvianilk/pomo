package main

import (
	"os"

	server "github.com/prithvianilk/pomo/internal/server"
)

func main() {
	server := server.New(os.Getenv("DB_URL"), os.Getenv("PORT"))
	server.Run()
}
