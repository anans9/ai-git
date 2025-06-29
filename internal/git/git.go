package git

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Client represents a Git client for repository operations
type Client struct {
	repo     *git.Repository
	workTree *git.Worktree
	repoPath string
}

// Status represents the status of files in the repository
type Status struct {
	Staged    []FileStatus
	Modified  []FileStatus
	Untracked []FileStatus
	Deleted   []FileStatus
	Renamed   []FileStatus
}

// FileStatus represents the status of a single file
type FileStatus struct {
	Path     string
	Status   string
	Original string // For renamed files
}

// Diff represents a git diff
type Diff struct {
	Files []FileDiff
	Stats DiffStats
}

// FileDiff represents changes to a single file
type FileDiff struct {
	Path      string
	OldPath   string
	Status    string
	Additions int
	Deletions int
	Content   string
}

// DiffStats represents statistics about a diff
type DiffStats struct {
	Files     int
	Additions int
	Deletions int
}

// Branch represents a git branch
type Branch struct {
	Name       string
	Current    bool
	Remote     string
	Upstream   string
	LastCommit string
}

// Commit represents a git commit
type Commit struct {
	Hash      string
	Message   string
	Author    string
	Email     string
	Date      time.Time
	ShortHash string
}

// Remote represents a git remote
type Remote struct {
	Name string
	URL  string
}

// NewClient creates a new Git client
func NewClient(path string) (*Client, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Find the git repository root
	repoPath, err := findGitRepo(path)
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	// Open the repository
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository: %w", err)
	}

	// Get the work tree
	workTree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get work tree: %w", err)
	}

	return &Client{
		repo:     repo,
		workTree: workTree,
		repoPath: repoPath,
	}, nil
}

// findGitRepo finds the git repository root starting from the given path
func findGitRepo(startPath string) (string, error) {
	path := startPath
	for {
		gitDir := filepath.Join(path, ".git")
		if info, err := os.Stat(gitDir); err == nil {
			if info.IsDir() {
				return path, nil
			}
			// Handle .git file (for worktrees)
			if content, err := os.ReadFile(gitDir); err == nil {
				if strings.HasPrefix(string(content), "gitdir:") {
					return path, nil
				}
			}
		}

		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "", fmt.Errorf("not a git repository")
}

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo(path string) bool {
	_, err := findGitRepo(path)
	return err == nil
}

// GetStatus returns the current status of the repository
func (c *Client) GetStatus() (*Status, error) {
	status, err := c.workTree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	result := &Status{
		Staged:    []FileStatus{},
		Modified:  []FileStatus{},
		Untracked: []FileStatus{},
		Deleted:   []FileStatus{},
		Renamed:   []FileStatus{},
	}

	for path, fileStatus := range status {
		fs := FileStatus{
			Path:   path,
			Status: string(fileStatus.Staging) + string(fileStatus.Worktree),
		}

		// Categorize based on status
		switch {
		case fileStatus.Staging != git.Unmodified:
			result.Staged = append(result.Staged, fs)
		case fileStatus.Worktree == git.Modified:
			result.Modified = append(result.Modified, fs)
		case fileStatus.Worktree == git.Untracked:
			result.Untracked = append(result.Untracked, fs)
		case fileStatus.Worktree == git.Deleted:
			result.Deleted = append(result.Deleted, fs)
		case fileStatus.Staging == git.Renamed:
			result.Renamed = append(result.Renamed, fs)
		}
	}

	return result, nil
}

// HasChanges checks if there are any changes in the repository
func (c *Client) HasChanges() (bool, error) {
	status, err := c.GetStatus()
	if err != nil {
		return false, err
	}

	return len(status.Modified) > 0 || len(status.Untracked) > 0 || len(status.Deleted) > 0 || len(status.Renamed) > 0, nil
}

// HasStagedChanges checks if there are any staged changes
func (c *Client) HasStagedChanges() (bool, error) {
	status, err := c.GetStatus()
	if err != nil {
		return false, err
	}

	return len(status.Staged) > 0, nil
}

// GetDiff returns the diff for unstaged changes
func (c *Client) GetDiff() (*Diff, error) {
	return c.getDiff(false)
}

// GetStagedDiff returns the diff for staged changes
func (c *Client) GetStagedDiff() (*Diff, error) {
	return c.getDiff(true)
}

func (c *Client) getDiff(staged bool) (*Diff, error) {
	head, err := c.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := c.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	var compareTree *object.Tree
	if staged {
		// Compare staged changes (index vs HEAD)
		// This is more complex and would require lower-level git operations
		// For now, we'll use a simplified approach
		compareTree = headTree
	} else {
		// Compare working directory vs HEAD
		compareTree = headTree
	}

	// Get file changes
	status, err := c.workTree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	diff := &Diff{
		Files: []FileDiff{},
		Stats: DiffStats{},
	}

	for filePath, fileStatus := range status {
		if staged && fileStatus.Staging == git.Unmodified {
			continue
		}
		if !staged && fileStatus.Worktree == git.Unmodified {
			continue
		}

		fileDiff := FileDiff{
			Path:   filePath,
			Status: string(fileStatus.Worktree),
		}

		// Try to get the actual diff content
		if content, err := c.getFileDiffContent(filePath, compareTree); err == nil {
			fileDiff.Content = content
			// Simple line counting (this could be improved)
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
					fileDiff.Additions++
				} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
					fileDiff.Deletions++
				}
			}
		}

		diff.Files = append(diff.Files, fileDiff)
		diff.Stats.Files++
		diff.Stats.Additions += fileDiff.Additions
		diff.Stats.Deletions += fileDiff.Deletions
	}

	return diff, nil
}

func (c *Client) getFileDiffContent(filePath string, compareTree *object.Tree) (string, error) {
	// Get file from working directory
	workingFile := filepath.Join(c.repoPath, filePath)
	workingContent, err := os.ReadFile(workingFile)
	if err != nil {
		return "", err
	}

	// Get file from tree (if it exists)
	var treeContent []byte
	if file, err := compareTree.File(filePath); err == nil {
		if reader, err := file.Reader(); err == nil {
			defer reader.Close()
			if content, err := io.ReadAll(reader); err == nil {
				treeContent = content
			}
		}
	}

	// Simple diff representation (this could be improved with a proper diff algorithm)
	if len(treeContent) == 0 {
		return fmt.Sprintf("+++ %s\n%s", filePath, string(workingContent)), nil
	}

	return fmt.Sprintf("--- %s\n+++ %s\n%s", filePath, filePath, string(workingContent)), nil
}

// Add stages files for commit
func (c *Client) Add(files ...string) error {
	if len(files) == 0 {
		// Add all files
		return c.workTree.AddWithOptions(&git.AddOptions{
			All: true,
		})
	}

	// Add specific files
	for _, file := range files {
		if err := c.workTree.AddWithOptions(&git.AddOptions{
			Path: file,
		}); err != nil {
			return fmt.Errorf("failed to add file %s: %w", file, err)
		}
	}

	return nil
}

// Commit creates a new commit with the given message
func (c *Client) Commit(message string) (*Commit, error) {
	// Get current user info
	cfg, err := c.repo.Config()
	if err != nil {
		return nil, fmt.Errorf("failed to get git config: %w", err)
	}

	var name, email string
	if cfg.User.Name != "" {
		name = cfg.User.Name
	} else {
		name = "Unknown"
	}
	if cfg.User.Email != "" {
		email = cfg.User.Email
	} else {
		email = "unknown@example.com"
	}

	// Create commit
	hash, err := c.workTree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  name,
			Email: email,
			When:  time.Now(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create commit: %w", err)
	}

	return &Commit{
		Hash:      hash.String(),
		ShortHash: hash.String()[:7],
		Message:   message,
		Author:    name,
		Email:     email,
		Date:      time.Now(),
	}, nil
}

// Push pushes commits to the remote repository
func (c *Client) Push() error {
	return c.repo.Push(&git.PushOptions{})
}

// Pull pulls changes from the remote repository
func (c *Client) Pull() error {
	return c.workTree.Pull(&git.PullOptions{})
}

// GetCurrentBranch returns the current branch name
func (c *Client) GetCurrentBranch() (string, error) {
	head, err := c.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}

	return head.Hash().String()[:7], nil // Return short hash if detached HEAD
}

// GetBranches returns all branches
func (c *Client) GetBranches() ([]Branch, error) {
	branches := []Branch{}

	// Get current branch
	currentBranch, _ := c.GetCurrentBranch()

	// Get local branches
	refs, err := c.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()
		branch := Branch{
			Name:    branchName,
			Current: branchName == currentBranch,
		}

		// Get last commit hash
		if commit, err := c.repo.CommitObject(ref.Hash()); err == nil {
			branch.LastCommit = commit.Hash.String()[:7]
		}

		branches = append(branches, branch)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to iterate branches: %w", err)
	}

	return branches, nil
}

// CreateBranch creates a new branch
func (c *Client) CreateBranch(name string) error {
	head, err := c.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	ref := plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), head.Hash())
	return c.repo.Storer.SetReference(ref)
}

// CheckoutBranch switches to the specified branch
func (c *Client) CheckoutBranch(name string) error {
	return c.workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
}

// GetRemotes returns all remotes
func (c *Client) GetRemotes() ([]Remote, error) {
	remotes, err := c.repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	result := []Remote{}
	for _, remote := range remotes {
		cfg := remote.Config()
		if len(cfg.URLs) > 0 {
			result = append(result, Remote{
				Name: cfg.Name,
				URL:  cfg.URLs[0],
			})
		}
	}

	return result, nil
}

// GetLastCommit returns the last commit
func (c *Client) GetLastCommit() (*Commit, error) {
	head, err := c.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commit, err := c.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}

	return &Commit{
		Hash:      commit.Hash.String(),
		ShortHash: commit.Hash.String()[:7],
		Message:   commit.Message,
		Author:    commit.Author.Name,
		Email:     commit.Author.Email,
		Date:      commit.Author.When,
	}, nil
}

// GetCommitHistory returns the commit history
func (c *Client) GetCommitHistory(limit int) ([]Commit, error) {
	head, err := c.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	commits := []Commit{}
	iter, err := c.repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer iter.Close()

	count := 0
	err = iter.ForEach(func(commit *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached") // Use error to break iteration
		}

		commits = append(commits, Commit{
			Hash:      commit.Hash.String(),
			ShortHash: commit.Hash.String()[:7],
			Message:   strings.TrimSpace(commit.Message),
			Author:    commit.Author.Name,
			Email:     commit.Author.Email,
			Date:      commit.Author.When,
		})

		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	return commits, nil
}

// IsClean checks if the working directory is clean
func (c *Client) IsClean() (bool, error) {
	status, err := c.workTree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	return status.IsClean(), nil
}

// GetRepoPath returns the repository path
func (c *Client) GetRepoPath() string {
	return c.repoPath
}

// GetChangedFiles returns a list of changed files
func (c *Client) GetChangedFiles() ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, file := range status.Modified {
		files = append(files, file.Path)
	}
	for _, file := range status.Untracked {
		files = append(files, file.Path)
	}
	for _, file := range status.Deleted {
		files = append(files, file.Path)
	}
	for _, file := range status.Renamed {
		files = append(files, file.Path)
	}

	return files, nil
}

// GetStagedFiles returns a list of staged files
func (c *Client) GetStagedFiles() ([]string, error) {
	status, err := c.GetStatus()
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, file := range status.Staged {
		files = append(files, file.Path)
	}

	return files, nil
}
