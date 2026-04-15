package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type DeploymentRepository struct {
	db *sqlx.DB
}

func NewDeploymentRepository(db *sqlx.DB) *DeploymentRepository {
	return &DeploymentRepository{db: db}
}

func (r *DeploymentRepository) List(stackID string, limit int) ([]models.Deployment, error) {
	var deployments []models.Deployment
	query := "SELECT * FROM deployments"
	args := []interface{}{}

	if stackID != "" {
		query += " WHERE stack_id = ?"
		args = append(args, stackID)
	}

	query += " ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	err := r.db.Select(&deployments, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}
	return deployments, nil
}

func (r *DeploymentRepository) GetByID(id string) (*models.Deployment, error) {
	var deployment models.Deployment
	err := r.db.Get(&deployment, "SELECT * FROM deployments WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("get deployment %s: %w", id, err)
	}
	return &deployment, nil
}

func (r *DeploymentRepository) Create(d *models.Deployment) error {
	_, err := r.db.NamedExec(`
		INSERT INTO deployments (id, stack_id, commit_hash, previous_commit, status, trigger_type, started_at, completed_at, error_message, created_at)
		VALUES (:id, :stack_id, :commit_hash, :previous_commit, :status, :trigger_type, :started_at, :completed_at, :error_message, :created_at)
	`, d)
	if err != nil {
		return fmt.Errorf("create deployment: %w", err)
	}
	return nil
}

func (r *DeploymentRepository) UpdateStatus(id, status, errorMessage string) error {
	_, err := r.db.Exec(`
		UPDATE deployments SET status = ?, error_message = ?,
		completed_at = CASE WHEN ? IN ('succeeded','failed','rolled_back') THEN datetime('now') ELSE completed_at END
		WHERE id = ?
	`, status, errorMessage, status, id)
	if err != nil {
		return fmt.Errorf("update deployment status: %w", err)
	}
	return nil
}
