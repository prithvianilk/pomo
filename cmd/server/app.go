package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	constants "github.com/prithvianilk/pomo/internal/constants"
	models "github.com/prithvianilk/pomo/internal/models"
)

const defaultStartDate = "2022-Sep-19"

type App struct {
	db *sqlx.DB
}

func (app *App) GetHandler(c *gin.Context) {
	startDate, endDate := getDateRanges(c)

	query := `SELECT * FROM session WHERE date BETWEEN $1 AND $2;`
	rows, err := app.db.Query(query, startDate, endDate)
	if err != nil {
		log.Printf("error during sql query: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	sessions, err := readSessions(rows)
	if err != nil {
		log.Printf("error while reading sessions: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	totalDuration := calculateTotalDuration(sessions)

	c.JSON(http.StatusOK, gin.H{
		"sessions":      sessions,
		"totalDuration": totalDuration,
	})
}

func (app *App) GetSingleSessionHandler(c *gin.Context) {
	name := c.Param("name")
	startDate, endDate := getDateRanges(c)

	query := `SELECT * FROM session 
	WHERE name = $1 AND 
	date BETWEEN $2 AND $3;`
	rows, err := app.db.Query(query, name, startDate, endDate)
	if err != nil {
		log.Printf("error during sql query: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	sessions, err := readSessions(rows)
	if err != nil {
		log.Printf("error while reading sessions: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	} else if len(sessions) == 0 {
		log.Printf("no sessions with name: %v", name)
		c.JSON(http.StatusNotFound, nil)
		return
	}

	totalDuration := calculateTotalDuration(sessions)

	c.JSON(http.StatusOK, gin.H{
		"sessions":      sessions,
		"totalDuration": totalDuration,
	})
}

func (app *App) PostHandler(c *gin.Context) {
	var session models.Session
	err := c.BindJSON(&session)
	if err != nil {
		log.Printf("error while parsing body: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	query := `INSERT INTO session (name, duration_in_minutes) VALUES ($1, $2);`
	_, err = app.db.Exec(query, session.Name, session.DurationInMinutes)
	if err != nil {
		log.Printf("error while inserting session: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (app *App) GetSessionNamesHandler(c *gin.Context) {
	query := `SELECT DISTINCT name from session;`
	rows, err := app.db.Query(query)
	if err != nil {
		log.Printf("error while querying session names: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	names, err := readNames(rows)
	if err != nil {
		log.Printf("error while reading session names: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, names)
}

func (app *App) DropAndCreateNew(c *gin.Context) {
	query := `DROP TABLE session;`
	_, err := app.db.Exec(query)
	if err != nil {
		log.Printf("error while dropping table: %v", err)
	}

	query = `CREATE TABLE session(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		date DATE NOT NULL DEFAULT CURRENT_DATE,
		duration_in_minutes INT
	);`
	_, err = app.db.Exec(query)
	if err != nil {
		log.Printf("error while creating table: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func getDateRanges(c *gin.Context) (string, string) {
	defaultEndDate := time.Now().Format(constants.DateLayout)
	startDate, endDate := c.DefaultQuery("start-date", defaultStartDate), c.DefaultQuery("end-date", defaultEndDate)
	return startDate, endDate
}

func readSessions(rows *sql.Rows) ([]models.Session, error) {
	var sessions []models.Session
	for rows.Next() {
		var session models.Session
		err := rows.Scan(&session.Id, &session.Name, &session.Date, &session.DurationInMinutes)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func readNames(rows *sql.Rows) ([]string, error) {
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func calculateTotalDuration(sessions []models.Session) int {
	total := 0
	for _, session := range sessions {
		total += session.DurationInMinutes
	}
	return total
}
