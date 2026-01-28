package models

import "github.com/google/uuid"

type Employee struct {
	ID         uuid.UUID
	FirstName  string
	LastName   string
	Email      string
	Department string
}
