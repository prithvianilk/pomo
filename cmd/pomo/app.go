package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/prithvianilk/pomo/internal/constants"
	"github.com/prithvianilk/pomo/internal/models"
)

const (
	spinChars             = `|/-\`
	serverConnFailMessage = "Error: pomo failed to connect to pomo-server. Maybe it's not running?"
)

var header = table.Row{"#", "Name", "Date", "Duration (M)"}

type SessionData struct {
	Sessions      []models.Session `json:"sessions"`
	TotalDuration int              `json:"totalDuration"`
}

type App struct {
	baseURL string
	flags   map[string]string
}

func (app *App) listSessions() {
	startDate, endDate := app.flags["start-date"], app.flags["end-date"]
	url := getRecordSessionURL(app.baseURL, startDate, endDate)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(serverConnFailMessage)
		return
	} else if resp.StatusCode == http.StatusNotFound {
		sessionName := flag.Args()[1]
		fmt.Printf("Error: There are no sessions with name: %v\n", sessionName)
		return
	}
	var sessionData SessionData
	err = json.NewDecoder(resp.Body).Decode(&sessionData)
	if err != nil {
		fmt.Printf("Error: Failed to decoding session data: %v", err)
		return
	}
	defer resp.Body.Close()
	printTable(sessionData)
}

func (app *App) recordSession() {
	name, duration := flag.Args()[1], flag.Args()[2]
	durationInMinutes, err := strconv.Atoi(duration)
	if err != nil {
		fmt.Printf("Incorrect argument: %s is not a number\n", duration)
		return
	}
	app.writePomoTickToStdout(durationInMinutes)
	session := models.Session{Name: name, DurationInMinutes: durationInMinutes}
	buff, _ := json.Marshal(session)
	body := bytes.NewBuffer(buff)
	url := app.baseURL + "/session"
	_, err = http.Post(url, "application/json", body)
	if err != nil {
		fmt.Println(serverConnFailMessage)
	}
}

func (a *App) parseFlags() {
	startDate, endDate := flag.String("start-date", "", ""), flag.String("end-date", "", "")
	flag.Parse()
	a.flags["start-date"] = *startDate
	a.flags["end-date"] = *endDate
}

func getRecordSessionURL(baseURL, startDate, endDate string) string {
	url := baseURL + "/session"
	isSingle := len(flag.Args()) > 1
	if isSingle {
		sessionName := flag.Args()[1]
		url += "/" + sessionName
	}
	if startDate == "" && endDate == "" {
		return url
	} else if startDate != "" && endDate != "" {
		return url + "?start-date=" + startDate + "&end-date=" + endDate
	} else if startDate != "" {
		return url + "?start-date=" + startDate
	}
	return url + "?end-date=" + endDate
}

func printTable(sessionData SessionData) {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.AppendHeader(header)
	for _, session := range sessionData.Sessions {
		writer.AppendRows([]table.Row{
			{session.Id, session.Name, session.Date.Format(constants.DateLayout), session.DurationInMinutes},
		})
	}
	writer.AppendFooter(table.Row{"", "", "Total", sessionData.TotalDuration})
	writer.Render()
}

func (*App) writePomoTickToStdout(durationInMinutes int) {
	durationInSec := durationInMinutes * 60
	for i := 0; i < durationInSec; i++ {
		spinChar := string(spinChars[(i % 4)])
		percentageDone := float64((i+1)*100) / float64(durationInSec)
		currentFormattedDuration := formatDuration(i)
		fmt.Printf("\r %s\tTime Elapsed: %s     Percentage Done: %.0f%s ", spinChar, currentFormattedDuration, percentageDone, "%")
		time.Sleep(time.Second)
	}
	fmt.Print("\n\n")
}

func formatDuration(duration int) string {
	durationInMinutes := int(math.Floor(float64(duration) / 60))
	if durationInMinutes == 0 {
		return formatDurationInSeconds(duration)
	}
	remainingSeconds := int(duration % 60)
	return fmt.Sprintf("%d Min/s, ", durationInMinutes) + formatDurationInSeconds(remainingSeconds)
}

func formatDurationInSeconds(duration int) string {
	if duration < 10 {
		return fmt.Sprintf(" %d Sec/s", duration)
	}
	return fmt.Sprintf("%d Sec/s", duration)
}
