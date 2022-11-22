package service

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prithvianilk/pomo/internal/types"
)

type PomoService struct {
	db *sqlx.DB
}

func getDB(dbURL string) *sqlx.DB {
	db, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	return db
}

func New(dbURL string) PomoService {
	return PomoService{db: getDB(dbURL)}
}

func (service PomoService) GetAllSessions(startDate, endDate string) (*types.SessionData, error) {
	query := `SELECT * FROM session WHERE date BETWEEN $1 AND $2;`
	rows, err := service.db.Query(query, startDate, endDate)
	if err != nil {
		log.Printf("error during sql query: %v", err)
		return nil, err
	}

	sessions, err := readSessions(rows)
	if err != nil {
		log.Printf("error while reading sessions: %v", err)
		return nil, err
	}
	totalDuration := calculateTotalDuration(sessions)

	return &types.SessionData{Sessions: sessions, TotalDuration: totalDuration}, nil
}

func (service PomoService) GetAllSessionsByName(name, startDate, endDate string) (*types.SessionData, error) {
	query := `SELECT * FROM session 
	WHERE name = $1 AND 
	date BETWEEN $2 AND $3;`
	rows, err := service.db.Query(query, name, startDate, endDate)
	if err != nil {
		log.Printf("error during sql query: %v", err)
		return nil, err
	}

	sessions, err := readSessions(rows)
	if err != nil {
		log.Printf("error while reading sessions: %v", err)
		return nil, err
	}
	totalDuration := calculateTotalDuration(sessions)

	return &types.SessionData{Sessions: sessions, TotalDuration: totalDuration}, nil
}

func (service PomoService) CreateNewSession(session types.Session) error {
	query := `INSERT INTO session (name, duration_in_minutes) VALUES ($1, $2);`
	_, err := service.db.Exec(query, session.Name, session.DurationInMinutes)
	if err != nil {
		log.Printf("error while inserting session: %v", err)
		return err
	}
	return nil
}

func (service PomoService) GetAllSessionNames() ([]string, error) {
	query := `SELECT DISTINCT name from session;`
	rows, err := service.db.Query(query)
	if err != nil {
		log.Printf("error while querying session names: %v", err)
		return nil, err
	}

	names, err := readNames(rows)
	if err != nil {
		log.Printf("error while reading session names: %v", err)
		return nil, err
	}
	return names, nil
}

func (service PomoService) DeleteSessionById(id string) error {
	query := `DELETE FROM session where id = $1;`
	_, err := service.db.Exec(query, id)
	if err != nil {
		log.Printf("error while deleting session with id %v: %v", id, err)
		return err
	}
	return nil
}

func (service PomoService) DropAndCreateNew() error {
	query := `DROP TABLE session;`
	_, err := service.db.Exec(query)
	if err != nil {
		log.Printf("error while dropping table: %v", err)
		return err
	}

	query = `CREATE TABLE session(
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		date DATE NOT NULL DEFAULT CURRENT_DATE,
		duration_in_minutes INT
	);`
	_, err = service.db.Exec(query)
	if err != nil {
		log.Printf("error while creating table: %v", err)
		return err
	}
	return nil
}

func readSessions(rows *sql.Rows) ([]types.Session, error) {
	var sessions []types.Session
	for rows.Next() {
		var session types.Session
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

func calculateTotalDuration(sessions []types.Session) int {
	total := 0
	for _, session := range sessions {
		total += session.DurationInMinutes
	}
	return total
}
