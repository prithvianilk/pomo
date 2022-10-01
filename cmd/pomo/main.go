package main

import (
	"flag"
	"os"

	models "github.com/prithvianilk/pomo/internal/models"
)

type SessionData struct {
	Sessions      []models.Session `json:"sessions"`
	TotalDuration int              `json:"totalDuration"`
}

func main() {
	app := App{baseURL: os.Getenv("BASE_URL"), flags: make(map[string]string)}
	app.parseFlags()
	cmd := flag.Args()[0]
	switch cmd {
	case "list":
		app.listSessions()
	case "record":
		app.recordSession()
	}
}
