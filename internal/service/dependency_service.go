package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/axism/composarr/internal/models"
	"github.com/axism/composarr/internal/repository"
	"github.com/google/uuid"
)

const (
	DependencyTypeHard = "hard"
	DependencyTypeSoft = "soft"
)

type DependencyService struct {
	depRepo   *repository.DependencyRepository
	stackRepo *repository.StackRepository
	dockerSvc *DockerService
}

func NewDependencyService(
	depRepo *repository.DependencyRepository,
	stackRepo *repository.StackRepository,
	dockerSvc *DockerService,
) *DependencyService {
	return &DependencyService{
		depRepo:   depRepo,
		stackRepo: stackRepo,
		dockerSvc: dockerSvc,
	}
}

type DependencyNode struct {
	Stack        models.Stack          `json:"stack"`
	Dependencies []models.StackDependency `json:"dependencies"`
	Dependents   []models.StackDependency `json:"dependents"`
}

type DependencyGraph struct {
	Nodes []DependencyNode           `json:"nodes"`
	Edges []models.StackDependency   `json:"edges"`
}

// AddDependency creates a new dependency edge after validating that it doesn't
// introduce a cycle.
func (d *DependencyService) AddDependency(stackID, dependsOnID, depType string) (*models.StackDependency, error) {
	if stackID == "" || dependsOnID == "" {
		return nil, errors.New("stackId and dependsOnId are required")
	}
	if stackID == dependsOnID {
		return nil, errors.New("a stack cannot depend on itself")
	}
	if depType == "" {
		depType = DependencyTypeHard
	}

	if err := d.checkCycleWithNewEdge(stackID, dependsOnID); err != nil {
		return nil, err
	}

	dep := &models.StackDependency{
		ID:             uuid.New().String(),
		StackID:        stackID,
		DependsOnID:    dependsOnID,
		DependencyType: depType,
		CreatedAt:      models.Now(),
	}

	if err := d.depRepo.Create(dep); err != nil {
		return nil, err
	}
	return dep, nil
}

func (d *DependencyService) RemoveDependency(id string) error {
	return d.depRepo.Delete(id)
}

func (d *DependencyService) ListDependencies(stackID string) ([]models.StackDependency, error) {
	return d.depRepo.ListDependencies(stackID)
}

func (d *DependencyService) ListDependents(stackID string) ([]models.StackDependency, error) {
	return d.depRepo.ListDependents(stackID)
}

// GetGraph returns all stacks and their dependency edges for visualization.
func (d *DependencyService) GetGraph() (*DependencyGraph, error) {
	stacks, err := d.stackRepo.List()
	if err != nil {
		return nil, err
	}

	edges, err := d.depRepo.ListAll()
	if err != nil {
		return nil, err
	}

	nodes := make([]DependencyNode, 0, len(stacks))
	for _, s := range stacks {
		node := DependencyNode{
			Stack:        s,
			Dependencies: []models.StackDependency{},
			Dependents:   []models.StackDependency{},
		}
		for _, e := range edges {
			if e.StackID == s.ID {
				node.Dependencies = append(node.Dependencies, e)
			}
			if e.DependsOnID == s.ID {
				node.Dependents = append(node.Dependents, e)
			}
		}
		nodes = append(nodes, node)
	}

	return &DependencyGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// GetDeployOrder returns the list of stack IDs in the order they should be
// deployed, with dependencies first. Uses Kahn's topological sort.
func (d *DependencyService) GetDeployOrder(stackID string) ([]string, error) {
	// Start from stackID and walk dependencies (transitively)
	visited := map[string]bool{}
	order := []string{}

	var dfs func(id string) error
	dfs = func(id string) error {
		if visited[id] {
			return nil
		}
		visited[id] = true

		deps, err := d.depRepo.ListDependencies(id)
		if err != nil {
			return err
		}

		for _, dep := range deps {
			if err := dfs(dep.DependsOnID); err != nil {
				return err
			}
		}
		order = append(order, id)
		return nil
	}

	if err := dfs(stackID); err != nil {
		return nil, err
	}
	return order, nil
}

// PreDeployCheck verifies that all hard dependencies of stackID are running.
// Returns an error identifying the failed dependency, or nil if all are healthy.
func (d *DependencyService) PreDeployCheck(ctx context.Context, stackID string) error {
	deps, err := d.depRepo.ListDependencies(stackID)
	if err != nil {
		return err
	}

	for _, dep := range deps {
		if dep.DependencyType != DependencyTypeHard {
			continue
		}

		depStack, err := d.stackRepo.GetByID(dep.DependsOnID)
		if err != nil {
			return fmt.Errorf("dependency stack %s not found: %w", dep.DependsOnID, err)
		}

		containers, err := d.dockerSvc.Ps(ctx, depStack.Slug)
		if err != nil {
			return fmt.Errorf("failed to check dependency %s: %w", depStack.Name, err)
		}

		if len(containers) == 0 {
			return fmt.Errorf("dependency %q is not running (no containers)", depStack.Name)
		}

		for _, c := range containers {
			if strings.ToLower(c.State) != "running" {
				return fmt.Errorf("dependency %q has container %s in state %s", depStack.Name, c.Name, c.State)
			}
		}
	}

	return nil
}

// checkCycleWithNewEdge runs DFS from dependsOnID looking for a path back to
// stackID. If found, adding the edge would create a cycle.
func (d *DependencyService) checkCycleWithNewEdge(stackID, dependsOnID string) error {
	visited := map[string]bool{}
	var dfs func(current string) bool
	dfs = func(current string) bool {
		if current == stackID {
			return true
		}
		if visited[current] {
			return false
		}
		visited[current] = true

		deps, err := d.depRepo.ListDependencies(current)
		if err != nil {
			return false
		}
		for _, dep := range deps {
			if dfs(dep.DependsOnID) {
				return true
			}
		}
		return false
	}

	if dfs(dependsOnID) {
		return errors.New("adding this dependency would create a cycle")
	}
	return nil
}
