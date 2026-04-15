package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/axism/composarr/internal/models"
	"github.com/axism/composarr/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type StackService struct {
	repo      *repository.StackRepository
	dockerSvc *DockerService
	cfg       *config.Config
}

func NewStackService(repo *repository.StackRepository, dockerSvc *DockerService, cfg *config.Config) *StackService {
	return &StackService{
		repo:      repo,
		dockerSvc: dockerSvc,
		cfg:       cfg,
	}
}

type CreateStackRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	ComposeContent string `json:"composeContent"`
}

type UpdateStackRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AutoUpdate  *bool   `json:"autoUpdate,omitempty"`
}

func (s *StackService) List() ([]models.Stack, error) {
	return s.repo.List()
}

func (s *StackService) GetByID(id string) (*models.Stack, error) {
	return s.repo.GetByID(id)
}

func (s *StackService) Create(req CreateStackRequest) (*models.Stack, error) {
	slug := slugify(req.Name)

	// Create the stack's compose directory
	stackDir := filepath.Join(s.cfg.ReposDir, slug)
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		return nil, fmt.Errorf("create stack directory: %w", err)
	}

	// Write the compose file
	composePath := filepath.Join(stackDir, "docker-compose.yml")
	if err := os.WriteFile(composePath, []byte(req.ComposeContent), 0644); err != nil {
		return nil, fmt.Errorf("write compose file: %w", err)
	}

	now := time.Now().UTC()
	stack := &models.Stack{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		ComposePath: "docker-compose.yml",
		Status:      "stopped",
		AutoUpdate:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(stack); err != nil {
		// Clean up directory on failure
		os.RemoveAll(stackDir)
		return nil, err
	}

	log.Info().Str("id", stack.ID).Str("name", stack.Name).Msg("stack created")
	return stack, nil
}

func (s *StackService) Update(id string, req UpdateStackRequest) (*models.Stack, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		stack.Name = *req.Name
	}
	if req.Description != nil {
		stack.Description = *req.Description
	}
	if req.AutoUpdate != nil {
		stack.AutoUpdate = *req.AutoUpdate
	}
	stack.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(stack); err != nil {
		return nil, err
	}

	return stack, nil
}

func (s *StackService) Delete(ctx context.Context, id string) error {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	// Stop containers first
	if _, err := s.dockerSvc.Down(ctx, stack.Slug); err != nil {
		log.Warn().Err(err).Str("stack", stack.Slug).Msg("failed to stop stack during deletion")
	}

	// Remove the stack's compose directory
	stackDir := filepath.Join(s.cfg.ReposDir, stack.Slug)
	if err := os.RemoveAll(stackDir); err != nil {
		log.Warn().Err(err).Str("dir", stackDir).Msg("failed to remove stack directory")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	log.Info().Str("id", id).Str("name", stack.Name).Msg("stack deleted")
	return nil
}

func (s *StackService) GetCompose(id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	composePath := filepath.Join(s.cfg.ReposDir, stack.Slug, stack.ComposePath)
	content, err := os.ReadFile(composePath)
	if err != nil {
		return "", fmt.Errorf("read compose file: %w", err)
	}

	return string(content), nil
}

func (s *StackService) UpdateCompose(id string, content string) error {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	composePath := filepath.Join(s.cfg.ReposDir, stack.Slug, stack.ComposePath)
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write compose file: %w", err)
	}

	stack.UpdatedAt = time.Now().UTC()
	return s.repo.Update(stack)
}

func (s *StackService) Start(ctx context.Context, id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	output, err := s.dockerSvc.Up(ctx, stack.Slug)
	if err != nil {
		return "", err
	}

	s.repo.UpdateStatus(id, "running")
	return output, nil
}

func (s *StackService) Stop(ctx context.Context, id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	output, err := s.dockerSvc.Down(ctx, stack.Slug)
	if err != nil {
		return "", err
	}

	s.repo.UpdateStatus(id, "stopped")
	return output, nil
}

func (s *StackService) Restart(ctx context.Context, id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	output, err := s.dockerSvc.Restart(ctx, stack.Slug)
	if err != nil {
		return "", err
	}

	return output, nil
}

func (s *StackService) Status(ctx context.Context, id string) ([]ContainerStatus, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	containers, err := s.dockerSvc.Ps(ctx, stack.Slug)
	if err != nil {
		return nil, err
	}

	// Update aggregate stack status
	status := "stopped"
	if len(containers) > 0 {
		allRunning := true
		for _, c := range containers {
			if c.State != "running" {
				allRunning = false
				break
			}
		}
		if allRunning {
			status = "running"
		} else {
			status = "degraded"
		}
	}
	s.repo.UpdateStatus(id, status)

	return containers, nil
}

func (s *StackService) Logs(ctx context.Context, id string, service string, tail int) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	return s.dockerSvc.Logs(ctx, stack.Slug, service, tail)
}

var nonAlphaRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = nonAlphaRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
