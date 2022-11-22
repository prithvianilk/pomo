package server

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prithvianilk/pomo/internal/constants"
	"github.com/prithvianilk/pomo/internal/server/service"
	"github.com/prithvianilk/pomo/internal/types"
)

const defaultStartDate = "2022-Sep-19"

type Server struct {
	service service.PomoService
	port    string
}

func New(dbURL, port string) *Server {
	service := service.New(dbURL)
	server := Server{service: service, port: port}
	return &server
}

func (server Server) Run() {
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

func (server Server) getHandler(c *gin.Context) {
	startDate, endDate := getDateRanges(c)

	sessionData, err := server.service.GetAllSessions(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, sessionData)
}

func (server Server) getSingleSessionHandler(c *gin.Context) {
	name := c.Param("name")
	startDate, endDate := getDateRanges(c)

	sessionData, err := server.service.GetAllSessionsByName(name, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, sessionData)
}

func (server Server) postHandler(c *gin.Context) {
	var session types.Session
	err := c.BindJSON(&session)
	if err != nil {
		log.Printf("error while parsing body: %v", err)
		c.JSON(http.StatusBadRequest, nil)
		return
	}

	err = server.service.CreateNewSession(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (server Server) getSessionNamesHandler(c *gin.Context) {
	names, err := server.service.GetAllSessionNames()
	if err != nil {
		log.Printf("error while querying session names: %v", err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, names)
}

func (server Server) deleteSessionHandler(c *gin.Context) {
	id := c.Param("id")

	err := server.service.DeleteSessionById(id)
	if err != nil {
		log.Printf("error while deleting session with id %v: %v", id, err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (server Server) dropAndCreateNew(c *gin.Context) {
	err := server.service.DropAndCreateNew()
	if err != nil {
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
