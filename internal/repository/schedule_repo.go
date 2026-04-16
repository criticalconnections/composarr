package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type ScheduleRepository struct {
	db *sqlx.DB
}

func NewScheduleRepository(db *sqlx.DB) *ScheduleRepository {
	return &ScheduleRepository{db: db}
}

func (r *ScheduleRepository) List() ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.db.Select(&schedules, "SELECT * FROM schedules ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list schedules: %w", err)
	}
	return schedules, nil
}

func (r *ScheduleRepository) ListByStack(stackID string) ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.db.Select(&schedules, "SELECT * FROM schedules WHERE stack_id = ? ORDER BY created_at DESC", stackID)
	if err != nil {
		return nil, fmt.Errorf("list schedules by stack: %w", err)
	}
	return schedules, nil
}

func (r *ScheduleRepository) ListEnabled() ([]models.Schedule, error) {
	var schedules []models.Schedule
	err := r.db.Select(&schedules, "SELECT * FROM schedules WHERE enabled = 1")
	if err != nil {
		return nil, fmt.Errorf("list enabled schedules: %w", err)
	}
	return schedules, nil
}

func (r *ScheduleRepository) GetByID(id string) (*models.Schedule, error) {
	var s models.Schedule
	err := r.db.Get(&s, "SELECT * FROM schedules WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("get schedule %s: %w", id, err)
	}
	return &s, nil
}

func (r *ScheduleRepository) Create(s *models.Schedule) error {
	_, err := r.db.NamedExec(`
		INSERT INTO schedules (id, stack_id, name, cron_expr, duration, timezone, enabled, created_at, updated_at)
		VALUES (:id, :stack_id, :name, :cron_expr, :duration, :timezone, :enabled, :created_at, :updated_at)
	`, s)
	if err != nil {
		return fmt.Errorf("create schedule: %w", err)
	}
	return nil
}

func (r *ScheduleRepository) Update(s *models.Schedule) error {
	_, err := r.db.NamedExec(`
		UPDATE schedules SET
			name = :name,
			cron_expr = :cron_expr,
			duration = :duration,
			timezone = :timezone,
			enabled = :enabled,
			updated_at = :updated_at
		WHERE id = :id
	`, s)
	if err != nil {
		return fmt.Errorf("update schedule: %w", err)
	}
	return nil
}

func (r *ScheduleRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM schedules WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete schedule: %w", err)
	}
	return nil
}
