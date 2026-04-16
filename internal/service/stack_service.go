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
	gitSvc    *GitService
	cfg       *config.Config
}

func NewStackService(repo *repository.StackRepository, dockerSvc *DockerService, gitSvc *GitService, cfg *config.Config) *StackService {
	return &StackService{
		repo:      repo,
		dockerSvc: dockerSvc,
		gitSvc:    gitSvc,
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

type UpdateComposeRequest struct {
	Content       string `json:"content"`
	CommitMessage string `json:"commitMessage,omitempty"`
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

	// Initialize git repo
	if _, err := s.gitSvc.InitRepo(slug); err != nil {
		os.RemoveAll(stackDir)
		return nil, fmt.Errorf("init git repo: %w", err)
	}

	// Write the compose file and create the initial commit
	commitMsg := fmt.Sprintf("Create stack: %s", req.Name)
	if _, err := s.gitSvc.WriteAndCommit(slug, "docker-compose.yml", []byte(req.ComposeContent), commitMsg); err != nil {
		os.RemoveAll(stackDir)
		return nil, fmt.Errorf("commit initial compose: %w", err)
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

	// Remove the stack's compose directory (includes git repo)
	if err := s.gitSvc.DeleteRepo(stack.Slug); err != nil {
		log.Warn().Err(err).Str("slug", stack.Slug).Msg("failed to remove stack directory")
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

	content, err := s.gitSvc.GetCurrentFile(stack.Slug, stack.ComposePath)
	if err != nil {
		return "", fmt.Errorf("read compose file: %w", err)
	}

	return string(content), nil
}

// UpdateCompose writes a new compose file, commits it to git, and returns the new commit hash.
func (s *StackService) UpdateCompose(id string, req UpdateComposeRequest) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	message := req.CommitMessage
	if strings.TrimSpace(message) == "" {
		message = "Update compose file"
	}

	commitHash, err := s.gitSvc.WriteAndCommit(stack.Slug, stack.ComposePath, []byte(req.Content), message)
	if err != nil {
		return "", fmt.Errorf("commit compose file: %w", err)
	}

	stack.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(stack); err != nil {
		return "", err
	}

	return commitHash, nil
}

// Rollback creates a new commit with the contents from the target commit.
func (s *StackService) Rollback(id, targetCommitHash string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	commitHash, err := s.gitSvc.RollbackToCommit(stack.Slug, targetCommitHash, stack.ComposePath)
	if err != nil {
		return "", err
	}

	stack.UpdatedAt = time.Now().UTC()
	s.repo.Update(stack)

	return commitHash, nil
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

// SlugForID returns the slug for a stack ID. Used by handlers that work with version operations.
func (s *StackService) SlugForID(id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return stack.Slug, nil
}

// ComposePathForID returns the compose filename for a stack ID.
func (s *StackService) ComposePathForID(id string) (string, error) {
	stack, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return stack.ComposePath, nil
}

var nonAlphaRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	slug := strings.ToLower(name)
	slug = nonAlphaRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}
