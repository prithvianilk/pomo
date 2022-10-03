package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/prithvianilk/pomo/internal/constants"
	models "github.com/prithvianilk/pomo/internal/models"
)

var (
	header    = table.Row{"#", "Name", "Date", "Duration (M)"}
	spinChars = `|/-\`
)

type App struct {
	baseURL string
	flags   map[string]string
}

func (a *App) parseFlags() {
	startDate, endDate := flag.String("start-date", "", ""), flag.String("end-date", "", "")
	flag.Parse()
	a.flags["start-date"] = *startDate
	a.flags["end-date"] = *endDate
}

func (app *App) listSessions() {
	startDate, endDate := app.flags["start-date"], app.flags["end-date"]
	url := getRecordSessionURL(app.baseURL, startDate, endDate)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	} else if resp.StatusCode == http.StatusNotFound {
		sessionName := flag.Args()[1]
		fmt.Printf("There are no sessions with name: %v\n", sessionName)
		return
	}
	var sessionData SessionData
	err = json.NewDecoder(resp.Body).Decode(&sessionData)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	printTable(sessionData)
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

func (app *App) recordSession() {
	name, duration := flag.Args()[1], flag.Args()[2]
	durationInMinutes, err := strconv.Atoi(duration)
	if err != nil {
		panic(err)
	}
	app.writePomoTickToStdout(durationInMinutes)
	session := models.Session{Name: name, DurationInMinutes: durationInMinutes}
	buf, err := json.Marshal(session)
	if err != nil {
		panic(err)
	}
	body := bytes.NewBuffer(buf)
	url := app.baseURL + "/session"
	_, err = http.Post(url, "application/json", body)
	if err != nil {
		panic(err)
	}
}

func (*App) writePomoTickToStdout(durationInMinutes int) {
	durationInSec := durationInMinutes * 60
	for i := 0; i < durationInSec; i++ {
		spinChar := string(spinChars[(i % 4)])
		percentageDone := (float32((i + 1)) * 100) / float32(durationInSec)
		fmt.Printf("\r %s\tTime Elapsed: %ds\tPercentage Done: %.2f", spinChar, i, percentageDone)
		time.Sleep(time.Second)
	}
}
