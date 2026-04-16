package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type DeploymentLogRepository struct {
	db *sqlx.DB
}

func NewDeploymentLogRepository(db *sqlx.DB) *DeploymentLogRepository {
	return &DeploymentLogRepository{db: db}
}

func (r *DeploymentLogRepository) Create(l *models.DeploymentLog) error {
	_, err := r.db.NamedExec(`
		INSERT INTO deployment_logs (id, deployment_id, level, message, timestamp)
		VALUES (:id, :deployment_id, :level, :message, :timestamp)
	`, l)
	if err != nil {
		return fmt.Errorf("create deployment log: %w", err)
	}
	return nil
}

func (r *DeploymentLogRepository) ListByDeployment(deploymentID string) ([]models.DeploymentLog, error) {
	var logs []models.DeploymentLog
	err := r.db.Select(&logs, `
		SELECT * FROM deployment_logs
		WHERE deployment_id = ?
		ORDER BY timestamp ASC
	`, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("list deployment logs: %w", err)
	}
	return logs, nil
}
