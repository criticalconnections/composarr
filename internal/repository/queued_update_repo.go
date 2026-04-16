package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type QueuedUpdateRepository struct {
	db *sqlx.DB
}

func NewQueuedUpdateRepository(db *sqlx.DB) *QueuedUpdateRepository {
	return &QueuedUpdateRepository{db: db}
}

func (r *QueuedUpdateRepository) List(stackID string) ([]models.QueuedUpdate, error) {
	var updates []models.QueuedUpdate
	var err error
	if stackID != "" {
		err = r.db.Select(&updates, "SELECT * FROM queued_updates WHERE stack_id = ? ORDER BY queued_at ASC", stackID)
	} else {
		err = r.db.Select(&updates, "SELECT * FROM queued_updates ORDER BY queued_at ASC")
	}
	if err != nil {
		return nil, fmt.Errorf("list queued updates: %w", err)
	}
	return updates, nil
}

func (r *QueuedUpdateRepository) ListByStackAndStatus(stackID, status string) ([]models.QueuedUpdate, error) {
	var updates []models.QueuedUpdate
	err := r.db.Select(&updates, `
		SELECT * FROM queued_updates
		WHERE stack_id = ? AND status = ?
		ORDER BY queued_at ASC
	`, stackID, status)
	if err != nil {
		return nil, fmt.Errorf("list queued updates by status: %w", err)
	}
	return updates, nil
}

func (r *QueuedUpdateRepository) GetByID(id string) (*models.QueuedUpdate, error) {
	var u models.QueuedUpdate
	err := r.db.Get(&u, "SELECT * FROM queued_updates WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("get queued update %s: %w", id, err)
	}
	return &u, nil
}

func (r *QueuedUpdateRepository) Create(u *models.QueuedUpdate) error {
	_, err := r.db.NamedExec(`
		INSERT INTO queued_updates (id, stack_id, schedule_id, compose_content, commit_message, status, queued_at, deploy_after, deployed_at)
		VALUES (:id, :stack_id, :schedule_id, :compose_content, :commit_message, :status, :queued_at, :deploy_after, :deployed_at)
	`, u)
	if err != nil {
		return fmt.Errorf("create queued update: %w", err)
	}
	return nil
}

func (r *QueuedUpdateRepository) UpdateStatus(id, status string) error {
	var err error
	if status == "deployed" {
		_, err = r.db.Exec("UPDATE queued_updates SET status = ?, deployed_at = datetime('now') WHERE id = ?", status, id)
	} else {
		_, err = r.db.Exec("UPDATE queued_updates SET status = ? WHERE id = ?", status, id)
	}
	if err != nil {
		return fmt.Errorf("update queued update status: %w", err)
	}
	return nil
}

func (r *QueuedUpdateRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM queued_updates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete queued update: %w", err)
	}
	return nil
}
