package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/axism/composarr/internal/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitInfo represents metadata about a git commit.
type CommitInfo struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Message   string    `json:"message"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

// GitService manages per-stack git repositories for compose file versioning.
type GitService struct {
	cfg *config.Config
}

func NewGitService(cfg *config.Config) *GitService {
	return &GitService{cfg: cfg}
}

const (
	authorName  = "Composarr"
	authorEmail = "composarr@localhost"
)

// repoPath returns the filesystem path for a stack's git repo.
func (g *GitService) repoPath(stackSlug string) string {
	return filepath.Join(g.cfg.ReposDir, stackSlug)
}

// InitRepo initializes a new git repository for a stack.
// The directory must already exist (created by StackService).
func (g *GitService) InitRepo(stackSlug string) (*git.Repository, error) {
	path := g.repoPath(stackSlug)
	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("init repo: %w", err)
	}
	return repo, nil
}

// openRepo opens an existing repository, returning an error if it doesn't exist.
func (g *GitService) openRepo(stackSlug string) (*git.Repository, error) {
	path := g.repoPath(stackSlug)
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("open repo %s: %w", stackSlug, err)
	}
	return repo, nil
}

// WriteAndCommit writes file content and creates a commit. Returns the commit hash.
// If there are no changes, returns the current HEAD hash without creating a commit.
func (g *GitService) WriteAndCommit(stackSlug, filename string, content []byte, message string) (string, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return "", err
	}

	// Write the file
	filePath := filepath.Join(g.repoPath(stackSlug), filename)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("get worktree: %w", err)
	}

	// Stage the file
	if _, err := wt.Add(filename); err != nil {
		return "", fmt.Errorf("stage file: %w", err)
	}

	// Check if there are actually changes to commit
	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("get status: %w", err)
	}
	if status.IsClean() {
		// No changes - return current HEAD
		return g.GetHeadCommit(stackSlug)
	}

	// Commit
	hash, err := wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	return hash.String(), nil
}

// GetCurrentFile reads the working-tree contents of a file.
func (g *GitService) GetCurrentFile(stackSlug, filename string) ([]byte, error) {
	filePath := filepath.Join(g.repoPath(stackSlug), filename)
	return os.ReadFile(filePath)
}

// GetFileAtCommit returns the contents of a file at a specific commit.
func (g *GitService) GetFileAtCommit(stackSlug, commitHash, filename string) ([]byte, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return nil, err
	}

	commit, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return nil, fmt.Errorf("get commit %s: %w", commitHash, err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get tree: %w", err)
	}

	file, err := tree.File(filename)
	if err != nil {
		return nil, fmt.Errorf("get file %s at %s: %w", filename, commitHash, err)
	}

	contents, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("read file contents: %w", err)
	}

	return []byte(contents), nil
}

// GetLog returns commit history for a stack, newest first.
func (g *GitService) GetLog(stackSlug string, limit int) ([]CommitInfo, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return []CommitInfo{}, nil
		}
		return nil, fmt.Errorf("get head: %w", err)
	}

	iter, err := repo.Log(&git.LogOptions{From: head.Hash()})
	if err != nil {
		return nil, fmt.Errorf("get log: %w", err)
	}
	defer iter.Close()

	var commits []CommitInfo
	count := 0
	err = iter.ForEach(func(c *object.Commit) error {
		if limit > 0 && count >= limit {
			return nil
		}
		hash := c.Hash.String()
		commits = append(commits, CommitInfo{
			Hash:      hash,
			ShortHash: hash[:8],
			Message:   c.Message,
			Author:    c.Author.Name,
			Email:     c.Author.Email,
			Timestamp: c.Author.When,
		})
		count++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("iterate log: %w", err)
	}

	return commits, nil
}

// GetCommitInfo returns metadata for a single commit.
func (g *GitService) GetCommitInfo(stackSlug, commitHash string) (*CommitInfo, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return nil, err
	}

	c, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return nil, fmt.Errorf("get commit: %w", err)
	}

	hash := c.Hash.String()
	return &CommitInfo{
		Hash:      hash,
		ShortHash: hash[:8],
		Message:   c.Message,
		Author:    c.Author.Name,
		Email:     c.Author.Email,
		Timestamp: c.Author.When,
	}, nil
}

// GetParentHash returns the hash of the immediate parent of a commit, or empty string for the root commit.
func (g *GitService) GetParentHash(stackSlug, commitHash string) (string, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return "", err
	}

	c, err := repo.CommitObject(plumbing.NewHash(commitHash))
	if err != nil {
		return "", err
	}

	if c.NumParents() == 0 {
		return "", nil
	}

	parent, err := c.Parent(0)
	if err != nil {
		return "", err
	}

	return parent.Hash.String(), nil
}

// GetHeadCommit returns the hash of the current HEAD.
func (g *GitService) GetHeadCommit(stackSlug string) (string, error) {
	repo, err := g.openRepo(stackSlug)
	if err != nil {
		return "", err
	}
	head, err := repo.Head()
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return "", nil
		}
		return "", err
	}
	return head.Hash().String(), nil
}

// DeleteRepo removes a stack's git repository.
func (g *GitService) DeleteRepo(stackSlug string) error {
	return os.RemoveAll(g.repoPath(stackSlug))
}

// RollbackToCommit rolls back the file to its state at the given commit by creating
// a new forward commit with the old contents. Returns the new commit hash.
func (g *GitService) RollbackToCommit(stackSlug, targetCommitHash, filename string) (string, error) {
	content, err := g.GetFileAtCommit(stackSlug, targetCommitHash, filename)
	if err != nil {
		return "", fmt.Errorf("read target commit: %w", err)
	}

	short := targetCommitHash
	if len(short) > 8 {
		short = short[:8]
	}
	message := fmt.Sprintf("Rollback to %s", short)
	return g.WriteAndCommit(stackSlug, filename, content, message)
}
