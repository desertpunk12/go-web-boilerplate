package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/repositories"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	err := config.LoadEnvFile()
	if err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	ctx := context.Background()
	dbInst, err := db.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer dbInst.Close()

	queries := repositories.New(dbInst.Pool)

	// Seed Users
	fmt.Println("Seeding 20 users...")
	for i := 0; i < 20; i++ {
		id := uuid.New()
		var pgID pgtype.UUID
		pgID.Bytes = id
		pgID.Valid = true

		_, err := queries.CreateUser(ctx, repositories.CreateUserParams{
			ID:       pgID,
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			Username: gofakeit.Username(),
			Password: gofakeit.Password(true, true, true, true, false, 12),
		})
		if err != nil {
			log.Printf("failed to create user %d: %v", i, err)
		}
	}

	// Seed Employees
	fmt.Println("Seeding 20 employees...")
	for i := 0; i < 20; i++ {
		id := uuid.New()
		var pgID pgtype.UUID
		pgID.Bytes = id
		pgID.Valid = true

		_, err := queries.CreateEmployee(ctx, repositories.CreateEmployeeParams{
			ID:         pgID,
			FirstName:  gofakeit.FirstName(),
			LastName:   gofakeit.LastName(),
			Email:      gofakeit.Email(),
			Department: gofakeit.JobTitle(),
		})
		if err != nil {
			log.Printf("failed to create employee %d: %v", i, err)
		}
	}

	fmt.Println("Seeding completed successfully!")
}
