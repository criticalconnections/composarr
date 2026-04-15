package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/axism/composarr/internal/config"
	"github.com/rs/zerolog/log"
)

// ContainerStatus represents the status of a container within a stack.
type ContainerStatus struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Service string `json:"service"`
	State   string `json:"state"`
	Status  string `json:"status"`
	Health  string `json:"health"`
}

// DockerService manages Docker compose operations by shelling out to the docker compose CLI.
type DockerService struct {
	cfg *config.Config
}

func NewDockerService(cfg *config.Config) *DockerService {
	return &DockerService{cfg: cfg}
}

func (d *DockerService) Up(ctx context.Context, stackSlug string) (string, error) {
	return d.runCompose(ctx, stackSlug, "up", "-d", "--remove-orphans")
}

func (d *DockerService) Down(ctx context.Context, stackSlug string) (string, error) {
	return d.runCompose(ctx, stackSlug, "down")
}

func (d *DockerService) Restart(ctx context.Context, stackSlug string) (string, error) {
	return d.runCompose(ctx, stackSlug, "restart")
}

func (d *DockerService) Pull(ctx context.Context, stackSlug string) (string, error) {
	return d.runCompose(ctx, stackSlug, "pull")
}

func (d *DockerService) Validate(ctx context.Context, stackSlug string) error {
	_, err := d.runCompose(ctx, stackSlug, "config", "--quiet")
	return err
}

func (d *DockerService) Ps(ctx context.Context, stackSlug string) ([]ContainerStatus, error) {
	output, err := d.runCompose(ctx, stackSlug, "ps", "--format", "json")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(output) == "" {
		return []ContainerStatus{}, nil
	}

	var containers []ContainerStatus
	// docker compose ps --format json outputs one JSON object per line
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw struct {
			ID      string `json:"ID"`
			Name    string `json:"Name"`
			Service string `json:"Service"`
			State   string `json:"State"`
			Status  string `json:"Status"`
			Health  string `json:"Health"`
		}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			log.Warn().Str("line", line).Err(err).Msg("failed to parse docker compose ps output")
			continue
		}
		containers = append(containers, ContainerStatus{
			ID:      raw.ID,
			Name:    raw.Name,
			Service: raw.Service,
			State:   raw.State,
			Status:  raw.Status,
			Health:  raw.Health,
		})
	}

	return containers, nil
}

func (d *DockerService) Logs(ctx context.Context, stackSlug string, service string, tail int) (string, error) {
	args := []string{"logs", "--no-color"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	if service != "" {
		args = append(args, service)
	}
	return d.runCompose(ctx, stackSlug, args...)
}

func (d *DockerService) runCompose(ctx context.Context, stackSlug string, args ...string) (string, error) {
	composeDir := filepath.Join(d.cfg.ReposDir, stackSlug)

	cmdArgs := append([]string{"compose", "-f", "docker-compose.yml", "-p", stackSlug}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Dir = composeDir
	cmd.Env = append(os.Environ(), fmt.Sprintf("COMPOSE_PROJECT_NAME=%s", stackSlug))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().
		Str("stack", stackSlug).
		Strs("args", cmdArgs).
		Msg("running docker compose")

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("docker compose %s: %s: %w", args[0], strings.TrimSpace(stderr.String()), err)
	}

	return stdout.String(), nil
}

// ComposeDir returns the filesystem path to a stack's compose directory.
func (d *DockerService) ComposeDir(stackSlug string) string {
	return filepath.Join(d.cfg.ReposDir, stackSlug)
}
