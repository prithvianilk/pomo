package main

import (
	"flag"
	"fmt"
	"os"

	models "github.com/prithvianilk/pomo/internal/models"
)

type SessionData struct {
	Sessions      []models.Session `json:"sessions"`
	TotalDuration int              `json:"totalDuration"`
}

func usage() {
	fmt.Println(`pomo: pomodoro timer and session manager for hackers.

COMMANDS
  pomo [--start-date=<value>] [--end-date=<value>] list                       List all pomo sessions.
  pomo [--start-date=<value>] [--end-date=<value>] list <name>                List all pomo sessions based on name.
  pomo record <name> <duration (M)>                                           Record a pomo session.
	`)
}

func main() {
	app := App{baseURL: os.Getenv("BASE_URL"), flags: make(map[string]string)}
	flag.Usage = usage
	app.parseFlags()
	if len(app.args) == 0 {
		flag.Usage()
		return
	}
	cmd := app.args[0]
	switch cmd {
	case "list":
		app.listSessions()
	case "record":
		app.recordSession()
	default:
		fmt.Printf("No command: %v\n", cmd)
		flag.Usage()
	}
}
