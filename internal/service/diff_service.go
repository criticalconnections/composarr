package service

import "errors"

// StructuredDiff is what the frontend uses to render side-by-side diffs.
type StructuredDiff struct {
	OldHash    string `json:"oldHash"`
	NewHash    string `json:"newHash"`
	OldContent string `json:"oldContent"`
	NewContent string `json:"newContent"`
}

// DiffService produces structured diffs between git versions.
type DiffService struct {
	gitSvc *GitService
}

func NewDiffService(gitSvc *GitService) *DiffService {
	return &DiffService{gitSvc: gitSvc}
}

// DiffVersions returns a structured diff between two commits.
// If fromHash is empty, it diffs the parent of toHash against toHash.
// If toHash is empty, it diffs fromHash against the working copy.
func (d *DiffService) DiffVersions(stackSlug, filename, fromHash, toHash string) (*StructuredDiff, error) {
	if fromHash == "" && toHash == "" {
		return nil, errors.New("at least one of fromHash or toHash must be provided")
	}

	var oldContent, newContent []byte
	var err error

	// Resolve fromHash content
	if fromHash == "" && toHash != "" {
		// Use parent of toHash as the "from"
		parentHash, perr := d.gitSvc.GetParentHash(stackSlug, toHash)
		if perr != nil {
			return nil, perr
		}
		fromHash = parentHash
	}

	if fromHash != "" {
		oldContent, err = d.gitSvc.GetFileAtCommit(stackSlug, fromHash, filename)
		if err != nil {
			// Parent might not exist (root commit) — that's fine, treat as empty
			oldContent = []byte("")
		}
	}

	// Resolve toHash content
	if toHash != "" {
		newContent, err = d.gitSvc.GetFileAtCommit(stackSlug, toHash, filename)
		if err != nil {
			return nil, err
		}
	} else {
		// Use working copy
		newContent, err = d.gitSvc.GetCurrentFile(stackSlug, filename)
		if err != nil {
			return nil, err
		}
	}

	return &StructuredDiff{
		OldHash:    fromHash,
		NewHash:    toHash,
		OldContent: string(oldContent),
		NewContent: string(newContent),
	}, nil
}

// DiffWorking returns the diff between HEAD and the working copy.
func (d *DiffService) DiffWorking(stackSlug, filename string) (*StructuredDiff, error) {
	headHash, err := d.gitSvc.GetHeadCommit(stackSlug)
	if err != nil {
		return nil, err
	}

	var oldContent []byte
	if headHash != "" {
		oldContent, err = d.gitSvc.GetFileAtCommit(stackSlug, headHash, filename)
		if err != nil {
			oldContent = []byte("")
		}
	}

	newContent, err := d.gitSvc.GetCurrentFile(stackSlug, filename)
	if err != nil {
		return nil, err
	}

	return &StructuredDiff{
		OldHash:    headHash,
		NewHash:    "",
		OldContent: string(oldContent),
		NewContent: string(newContent),
	}, nil
}
