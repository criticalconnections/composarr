package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type StackRepository struct {
	db *sqlx.DB
}

func NewStackRepository(db *sqlx.DB) *StackRepository {
	return &StackRepository{db: db}
}

func (r *StackRepository) List() ([]models.Stack, error) {
	var stacks []models.Stack
	err := r.db.Select(&stacks, "SELECT * FROM stacks ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("list stacks: %w", err)
	}
	return stacks, nil
}

func (r *StackRepository) GetByID(id string) (*models.Stack, error) {
	var stack models.Stack
	err := r.db.Get(&stack, "SELECT * FROM stacks WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("get stack %s: %w", id, err)
	}
	return &stack, nil
}

func (r *StackRepository) GetBySlug(slug string) (*models.Stack, error) {
	var stack models.Stack
	err := r.db.Get(&stack, "SELECT * FROM stacks WHERE slug = ?", slug)
	if err != nil {
		return nil, fmt.Errorf("get stack by slug %s: %w", slug, err)
	}
	return &stack, nil
}

func (r *StackRepository) Create(stack *models.Stack) error {
	_, err := r.db.NamedExec(`
		INSERT INTO stacks (id, name, slug, description, compose_path, status, auto_update, created_at, updated_at)
		VALUES (:id, :name, :slug, :description, :compose_path, :status, :auto_update, :created_at, :updated_at)
	`, stack)
	if err != nil {
		return fmt.Errorf("create stack: %w", err)
	}
	return nil
}

func (r *StackRepository) Update(stack *models.Stack) error {
	_, err := r.db.NamedExec(`
		UPDATE stacks SET
			name = :name,
			description = :description,
			auto_update = :auto_update,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id
	`, stack)
	if err != nil {
		return fmt.Errorf("update stack: %w", err)
	}
	return nil
}

func (r *StackRepository) UpdateStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE stacks SET status = ?, updated_at = datetime('now') WHERE id = ?", status, id)
	if err != nil {
		return fmt.Errorf("update stack status: %w", err)
	}
	return nil
}

func (r *StackRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM stacks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete stack: %w", err)
	}
	return nil
}
