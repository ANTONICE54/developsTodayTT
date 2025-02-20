package models

import "time"

type Cat struct {
	ID                uint      `json:"id"`
	Name              string    `json:"name"`
	YearsOfExperience uint      `json:"years_of_experience"`
	Breed             string    `json:"breed"`
	Salary            float64   `json:"salary"`
	CreatedAt         time.Time `json:"created_at"`
}
