package pkg

import "time"

type Session struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	Date              time.Time `json:"date"`
	DurationInMinutes int       `json:"duration_in_minutes"`
}
