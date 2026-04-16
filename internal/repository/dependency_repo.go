package repository

import (
	"fmt"

	"github.com/axism/composarr/internal/models"
	"github.com/jmoiron/sqlx"
)

type DependencyRepository struct {
	db *sqlx.DB
}

func NewDependencyRepository(db *sqlx.DB) *DependencyRepository {
	return &DependencyRepository{db: db}
}

// ListAll returns every dependency edge in the system.
func (r *DependencyRepository) ListAll() ([]models.StackDependency, error) {
	var deps []models.StackDependency
	err := r.db.Select(&deps, "SELECT * FROM stack_dependencies")
	if err != nil {
		return nil, fmt.Errorf("list all dependencies: %w", err)
	}
	return deps, nil
}

// ListDependencies returns stacks that stackID depends on.
func (r *DependencyRepository) ListDependencies(stackID string) ([]models.StackDependency, error) {
	var deps []models.StackDependency
	err := r.db.Select(&deps, "SELECT * FROM stack_dependencies WHERE stack_id = ?", stackID)
	if err != nil {
		return nil, fmt.Errorf("list dependencies: %w", err)
	}
	return deps, nil
}

// ListDependents returns stacks that depend on stackID.
func (r *DependencyRepository) ListDependents(stackID string) ([]models.StackDependency, error) {
	var deps []models.StackDependency
	err := r.db.Select(&deps, "SELECT * FROM stack_dependencies WHERE depends_on_id = ?", stackID)
	if err != nil {
		return nil, fmt.Errorf("list dependents: %w", err)
	}
	return deps, nil
}

func (r *DependencyRepository) Create(d *models.StackDependency) error {
	_, err := r.db.NamedExec(`
		INSERT INTO stack_dependencies (id, stack_id, depends_on_id, dependency_type, created_at)
		VALUES (:id, :stack_id, :depends_on_id, :dependency_type, :created_at)
	`, d)
	if err != nil {
		return fmt.Errorf("create dependency: %w", err)
	}
	return nil
}

func (r *DependencyRepository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM stack_dependencies WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete dependency: %w", err)
	}
	return nil
}

func (r *DependencyRepository) DeleteByStacks(stackID, dependsOnID string) error {
	_, err := r.db.Exec(
		"DELETE FROM stack_dependencies WHERE stack_id = ? AND depends_on_id = ?",
		stackID, dependsOnID,
	)
	if err != nil {
		return fmt.Errorf("delete dependency by stacks: %w", err)
	}
	return nil
}
