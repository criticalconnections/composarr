package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type HealthCheckRepository struct {
	db *sqlx.DB
}

func NewHealthCheckRepository(db *sqlx.DB) *HealthCheckRepository {
	return &HealthCheckRepository{db: db}
}

func (r *HealthCheckRepository) Create(h *models.HealthCheckResult) error {
	_, err := r.db.NamedExec(`
		INSERT INTO health_check_results (id, deployment_id, container_name, service_name, status, check_output, checked_at)
		VALUES (:id, :deployment_id, :container_name, :service_name, :status, :check_output, :checked_at)
	`, h)
	if err != nil {
		return fmt.Errorf("create health check result: %w", err)
	}
	return nil
}

// ListByDeployment returns all health-check rows for a deployment, oldest first.
func (r *HealthCheckRepository) ListByDeployment(deploymentID string) ([]models.HealthCheckResult, error) {
	var results []models.HealthCheckResult
	err := r.db.Select(&results, `
		SELECT * FROM health_check_results
		WHERE deployment_id = ?
		ORDER BY checked_at ASC
	`, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("list health check results: %w", err)
	}
	return results, nil
}

// LatestByDeployment returns the most recent health-check row for each container in a deployment.
func (r *HealthCheckRepository) LatestByDeployment(deploymentID string) ([]models.HealthCheckResult, error) {
	var results []models.HealthCheckResult
	err := r.db.Select(&results, `
		SELECT h.* FROM health_check_results h
		INNER JOIN (
			SELECT container_name, MAX(checked_at) AS latest
			FROM health_check_results
			WHERE deployment_id = ?
			GROUP BY container_name
		) latest_h ON h.container_name = latest_h.container_name AND h.checked_at = latest_h.latest
		WHERE h.deployment_id = ?
		ORDER BY h.service_name ASC
	`, deploymentID, deploymentID)
	if err != nil {
		return nil, fmt.Errorf("get latest health checks: %w", err)
	}
	return results, nil
}
