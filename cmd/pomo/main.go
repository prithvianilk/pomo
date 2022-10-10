package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Println(`pomo: pomodoro timer and session manager for hackers.

COMMANDS
  pomo [--start-date=<value>] [--end-date=<value>] list           List all pomo sessions.
  pomo [--start-date=<value>] [--end-date=<value>] list <name>    List all pomo sessions based on name.
  pomo --nameonly list                                            List all pomo session names.
  pomo record <name> <duration (M)>                               Record a pomo session.`)
}

func main() {
	app := App{baseURL: os.Getenv("BASE_URL"), Flags: Flags{}}
	flag.Usage = usage
	app.parseFlags()
	if len(flag.Args()) == 0 {
		flag.Usage()
		return
	}
	cmd := flag.Args()[0]
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
