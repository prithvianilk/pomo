package main

import (
	"flag"
	"fmt"
	"os"

	app "github.com/prithvianilk/pomo/internal/pomo"
)

func usage() {
	fmt.Println(`pomo: pomodoro timer and session manager for hackers.

COMMANDS
  pomo [--start-date=<value>] [--end-date=<value>] list           List all pomo sessions.
  pomo [--start-date=<value>] [--end-date=<value>] list <name>    List all pomo sessions based on name.
  pomo --nameonly list                                            List all pomo session names.
  pomo record <name> <duration (M)>                               Record a pomo session.
  pomo delete <id>                                                Delete a pomo session by id.`)
}

func main() {
	app := app.New(os.Getenv("BASE_URL"))
	app.ParseFlags()

	flag.Usage = usage
	if len(flag.Args()) == 0 {
		flag.Usage()
		return
	}

	cmd := flag.Args()[0]
	switch cmd {
	case "list":
		app.ListSessions()
	case "record":
		app.RecordSession()
	case "delete":
		app.DeleteSession()
	default:
		fmt.Printf("No command: %v\n", cmd)
		flag.Usage()
	}
}
