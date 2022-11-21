package app

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

	"github.com/gen2brain/beeep"
	"github.com/jedib0t/go-pretty/table"
	"github.com/prithvianilk/pomo/internal/constants"
	"github.com/prithvianilk/pomo/internal/models"
)

const (
	spinChars               = `|/-\`
	serverConnFailMessage   = "Error: pomo failed to connect to pomo-server. Maybe it's not running?"
	serverGenericErrMessage = "Unable to perform command. Some issue occured."
)

var header = table.Row{"#", "Name", "Date", "Duration (M)"}

type sessionData struct {
	Sessions      []models.Session `json:"sessions"`
	TotalDuration int              `json:"totalDuration"`
}

type flags struct {
	startDate, endDate string
	isNameOnlyCommand  bool
}

type App struct {
	baseURL string
	flags
}

func New(baseURL string) *App {
	app := App{baseURL: baseURL, flags: flags{}}
	return &app
}

func (app *App) ListSessions() {
	if app.isNameOnlyCommand {
		app.listSessionNames()
		return
	}

	url := getRecordSessionURL(app.baseURL, app.startDate, app.endDate)
	resp, err := http.Get(url)
	if checkAndHandleConnFailOrInternalServerError(err, resp) {
		return
	} else if resp.StatusCode == http.StatusNotFound {
		sessionName := flag.Args()[1]
		fmt.Printf("Error: There are no sessions with name: %v\n", sessionName)
		return
	}

	var sessionData sessionData
	err = json.NewDecoder(resp.Body).Decode(&sessionData)
	if err != nil {
		fmt.Printf("Error: Failed to decoding session data: %v", err)
		return
	}
	defer resp.Body.Close()
	printTable(sessionData)
}

func (app *App) RecordSession() {
	name, duration := flag.Args()[1], flag.Args()[2]
	durationInMinutes, err := strconv.Atoi(duration)
	if err != nil {
		fmt.Printf("Incorrect argument: %s is not a number\n", duration)
		return
	}
	writePomoTickToStdout(durationInMinutes)
	session := models.Session{Name: name, DurationInMinutes: durationInMinutes}
	buff, _ := json.Marshal(session)
	body := bytes.NewBuffer(buff)
	url := app.baseURL + "/session"
	resp, err := http.Post(url, "application/json", body)
	if err != nil {
		fmt.Println(serverConnFailMessage)
		return
	} else if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println(serverGenericErrMessage)
		return
	}
	notifyOnDesktop(name)
}

func (app *App) listSessionNames() {
	url := app.baseURL + "/name"
	resp, err := http.Get(url)
	if checkAndHandleConnFailOrInternalServerError(err, resp) {
		return
	}
	var names []string
	err = json.NewDecoder(resp.Body).Decode(&names)
	if err != nil {
		fmt.Printf("Error: Failed to decoding session names: %v", err)
		return
	}
	for _, name := range names {
		fmt.Println(name)
	}
}

func (a *App) ParseFlags() {
	isNameOnlyCommand := flag.Bool("nameonly", false, "")
	startDate, endDate := flag.String("start-date", "", ""), flag.String("end-date", "", "")
	flag.Parse()
	a.isNameOnlyCommand = *isNameOnlyCommand
	a.startDate, a.endDate = *startDate, *endDate
}

func checkAndHandleConnFailOrInternalServerError(err error, resp *http.Response) bool {
	if err != nil {
		fmt.Println(serverConnFailMessage)
		return true
	} else if resp.StatusCode == http.StatusInternalServerError {
		fmt.Println(serverGenericErrMessage)
		return true
	}
	return false
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

func printTable(sessionData sessionData) {
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

func writePomoTickToStdout(durationInMinutes int) {
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

func notifyOnDesktop(name string) {
	err := beeep.Notify("pomo", fmt.Sprintf("%s session completed!", name), "")
	if err != nil {
		panic(err)
	}
}
