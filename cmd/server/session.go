package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	models "github.com/prithvianilk/pomo/pkg/models"
)

func GetHandler(c *gin.Context) {
	db := ParseDB(c)
	startDate, endDate := c.Query("start-date"), c.Query("end-date")
	var (
		rows *sql.Rows
		err  error
	)
	isRangeQuery := startDate != "" && endDate != ""
	isInvalid := !isRangeQuery && (startDate != "" || endDate != "")
	if isInvalid {
		log.Printf("only one date range present: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	} else if !isRangeQuery {
		query := `SELECT * FROM session;`
		rows, err = db.Query(query)
	} else {
		query := `SELECT * FROM session WHERE date BETWEEN $1 AND $2;`
		rows, err = db.Query(query, startDate, endDate)
	}
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

func calculateTotalDuration(sessions []models.Session) int {
	total := 0
	for _, session := range sessions {
		total += session.DurationInMinutes
	}
	return total
}

func PostHandler(c *gin.Context) {
	db := ParseDB(c)
	var session models.Session
	err := c.BindJSON(&session)
	if err != nil {
		log.Printf("error while parsing body: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	query := `INSERT INTO session (name, duration_in_minutes) VALUES ($1, $2);`
	_, err = db.Exec(query, session.Name, session.DurationInMinutes)
	if err != nil {
		log.Printf("error while inserting session: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func DropAndCreateNew(c *gin.Context) {
	db := ParseDB(c)

	query := `DROP TABLE session;`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("error while dropping table: %v", err)
	}

	query = `CREATE TABLE session(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		date DATE NOT NULL DEFAULT CURRENT_DATE,
		duration_in_minutes INT
	);`
	_, err = db.Exec(query)
	if err != nil {
		log.Printf("error while creating table: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}
