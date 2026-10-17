package repository

import (
	"context"
	"taskqueue/internal/model"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepo struct {
	DB *pgxpool.Pool
}

func NewPostgesRepo(DB *pgxpool.Pool) *PostgresRepo {
	return &PostgresRepo{DB: DB}
}

func (repo *PostgresRepo) CreateTask(ctx context.Context, task *model.Task) (int64, error) {
	query := `INSERT INTO tasks ( name) VALUES ($1) RETURNING id;`

	var taskID int64
	err := repo.DB.QueryRow(ctx, query, task.Name).Scan(&taskID)

	if err != nil {
		return 0, err
	}

	return taskID, nil
}

func (repo *PostgresRepo) GetTask(ctx context.Context, taskID int64) (*model.Task, error) {
	query := `SELECT id,name,status,attempts,COALESCE(error, '') AS error,created_at,updated_at,lease_expires_at
			  FROM tasks WHERE id = $1`

	var task model.Task
	var LeaseExpiresAt pgtype.Timestamptz

	err := repo.DB.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.Name,
		&task.Status,
		&task.Attempts,
		&task.Error,
		&task.CreatedAt,
		&task.UpdatedAt,
		&LeaseExpiresAt)


	if LeaseExpiresAt.Valid {
		task.LeaseExpiresAt = &LeaseExpiresAt.Time
	} else {
		task.LeaseExpiresAt = nil
	}

	if err != nil {
		return &model.Task{}, err
	}

	return &task, nil
}

func (repo *PostgresRepo) UpdateTask(ctx context.Context, task *model.Task) error {
	query := `UPDATE tasks
				SET status = $1,
				attempts = $2,
				error = $3,
				updated_at = NOW(),
				lease_expires_at = $4 
				WHERE id = $5;`

	_, err := repo.DB.Exec(ctx, query, task.Status, task.Attempts, task.Error, task.LeaseExpiresAt, task.ID)

	if err != nil {
		return err
	}

	return nil
}

func (repo *PostgresRepo) GetPendingTasks(ctx context.Context) (*[]model.Pair, error) {

	query := `SELECT id,created_at FROM tasks WHERE status = $1`

	rows, err := repo.DB.Query(ctx, query, "pending")

	if err != nil {
		return &[]model.Pair{}, err
	}

	var response []model.Pair

	for rows.Next() {
		var pair model.Pair
		err := rows.Scan(&pair.TaskID, &pair.CreatedAt)

		if err != nil {
			return &[]model.Pair{}, err
		}

		response = append(response, pair)
	}

	defer rows.Close()

	return &response, nil

}

func (repo *PostgresRepo) RecoverExpiredTasks(ctx context.Context) (*[]int64, error) {
	query := `UPDATE tasks 
			SET status = $1,
				lease_expires_at = $2,
				updated_at = NOW()
			WHERE status = $3 
				AND lease_expires_at <= NOW() 
			RETURNING id;`

	rows, err := repo.DB.Query(ctx, query, "pending", nil, "pending")

	if err != nil {
		return &[]int64{}, err
	}

	var TaskIDs []int64

	for rows.Next() {
		var ID int64

		err := rows.Scan(&ID)

		if err != nil {
			return &[]int64{}, err
		}

		TaskIDs = append(TaskIDs, ID)
	}

	defer rows.Close()

	return &TaskIDs, nil

}

func (repo *PostgresRepo) ClaimTask(ctx context.Context, taskID int64) (int64, error) {

	query := `UPDATE tasks
				SET status = 'processing',
					attempts = attempts + 1,
					lease_expires_at = NOW() + INTERVAL '30 seconds',
					updated_at = NOW()
				WHERE id = $1
				AND status = 'pending'
				RETURNING id;`
	var claimedID int64
	err := repo.DB.QueryRow(ctx,query,taskID).Scan(&claimedID)

	if err != nil {
		return 0,err
	}

	return claimedID,err
	
}
