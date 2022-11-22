package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	constants "github.com/prithvianilk/pomo/internal/constants"
	models "github.com/prithvianilk/pomo/internal/models"
)

const defaultStartDate = "2022-Sep-19"

func getDB(dbURL string) *sqlx.DB {
	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	return db
}

type Server struct {
	db   *sqlx.DB
	port string
}

func New(dbURL, port string) *Server {
	server := Server{db: getDB(dbURL), port: port}
	return &server
}

func (server *Server) Run() {
	router := gin.Default()

	router.GET("/session", server.getHandler)
	router.GET("/session/:name", server.getSingleSessionHandler)
	router.POST("/session", server.postHandler)
	router.GET("/name", server.getSessionNamesHandler)
	router.DELETE("/session/:id", server.deleteSessionHandler)
	router.GET("/maintainance/session", server.dropAndCreateNew)

	address := ":" + server.port
	err := router.Run(address)
	if err != nil {
		log.Fatalf("error while running server: %v", err)
	}
}

func (server *Server) getHandler(c *gin.Context) {
	startDate, endDate := getDateRanges(c)

	query := `SELECT * FROM session WHERE date BETWEEN $1 AND $2;`
	rows, err := server.db.Query(query, startDate, endDate)
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

func (server *Server) getSingleSessionHandler(c *gin.Context) {
	name := c.Param("name")
	startDate, endDate := getDateRanges(c)

	query := `SELECT * FROM session 
	WHERE name = $1 AND 
	date BETWEEN $2 AND $3;`
	rows, err := server.db.Query(query, name, startDate, endDate)
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

func (server *Server) postHandler(c *gin.Context) {
	var session models.Session
	err := c.BindJSON(&session)
	if err != nil {
		log.Printf("error while parsing body: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}
	query := `INSERT INTO session (name, duration_in_minutes) VALUES ($1, $2);`
	_, err = server.db.Exec(query, session.Name, session.DurationInMinutes)
	if err != nil {
		log.Printf("error while inserting session: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (server *Server) getSessionNamesHandler(c *gin.Context) {
	query := `SELECT DISTINCT name from session;`
	rows, err := server.db.Query(query)
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

func (server *Server) deleteSessionHandler(c *gin.Context) {
	id := c.Param("id")
	query := `DELETE FROM session where id = $1;`
	_, err := server.db.Exec(query, id)
	if err != nil {
		log.Printf("error while deleting session with id %v: %v", id, err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (server *Server) dropAndCreateNew(c *gin.Context) {
	query := `DROP TABLE session;`
	_, err := server.db.Exec(query)
	if err != nil {
		log.Printf("error while dropping table: %v", err)
	}

	query = `CREATE TABLE session(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		date DATE NOT NULL DEFAULT CURRENT_DATE,
		duration_in_minutes INT
	);`
	_, err = server.db.Exec(query)
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
