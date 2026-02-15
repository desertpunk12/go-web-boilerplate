package main

import (
	"context"
	"os"
	"web-boilerplate/internal/hr-api/config"
	"web-boilerplate/internal/hr-api/db"
	"web-boilerplate/internal/hr-api/repositories"
	"web-boilerplate/shared/helpers"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	err := config.LoadEnvFile()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load env")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal().Msg("DATABASE_URL is not set")
	}

	ctx := context.Background()
	dbInst, err := db.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to db")
	}
	defer dbInst.Close()

	queries := repositories.New(dbInst.Pool)

	// Seed Users
	log.Info().Msg("Seeding 20 users...")
	for i := range 20 {
		id := uuid.New()
		var pgID pgtype.UUID
		pgID.Bytes = id
		pgID.Valid = true
		pass, err := helpers.HashPass(gofakeit.Password(true, true, true, true, false, 12))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to hash password")
		}
		_, err = queries.CreateUser(ctx, repositories.CreateUserParams{
			ID:       pgID,
			Name:     gofakeit.Name(),
			Email:    gofakeit.Email(),
			Username: gofakeit.Username(),
			Password: pass,
		})
		if err != nil {
			log.Error().Err(err).Int("user_index", i).Msg("failed to create user")
		}
	}

	// Seed Employees
	log.Info().Msg("Seeding 20 employees...")
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
			log.Error().Err(err).Int("employee_index", i).Msg("failed to create employee")
		}
	}

	log.Info().Msg("Seeding completed successfully!")
}
