package database

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"os"
)

func ConnectDB() (*pgxpool.Pool, error) {
	err := godotenv.Load()

	if err != nil {
		return nil, fmt.Errorf("failed to load env : %w", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	config, configErr := pgxpool.ParseConfig(dsn)

	if configErr != nil {
		return nil, err
	}

	pool, dbErr := pgxpool.NewWithConfig(context.Background(), config)

	if dbErr != nil {
		return nil, dbErr
	}

	poolErr := pool.Ping(context.Background())

	if poolErr != nil {
		return nil, poolErr
	}

	return pool, nil

}
