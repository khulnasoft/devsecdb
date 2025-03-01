package vcs

import (
	"context"
	"fmt"
	"sync"

	storepb "github.com/khulnasoft/devsecdb/proto/generated-go/store"
)

// RefType is the type of a ref.
type RefType string

const (
	// RefTypeBranch is the branch ref type.
	RefTypeBranch RefType = "branch"
	// RefTypeTag is the tag ref type.
	RefTypeTag RefType = "tag"
	// RefTypeCommit is the commit ref type.
	RefTypeCommit RefType = "commit"
)

// RefInfo is the API message for a VCS ref.
type RefInfo struct {
	RefType RefType
	RefName string
}

// FileDiffType is the type of file diff.
type FileDiffType int

const (
	// FileDiffTypeUnknown means the file is an unknown diff type.
	FileDiffTypeUnknown FileDiffType = iota
	// FileDiffTypeAdded means the file is newly added.
	FileDiffTypeAdded
	// FileDiffTypeModified means the file is modified.
	FileDiffTypeModified
	// FileDiffTypeRemoved means the file is removed.
	FileDiffTypeRemoved
)

// FileDiff contains file diffs between two commits.
// It's obtained by comparing the base and head commits of a PR/MR so that we know the real changes.
type FileDiff struct {
	Path string
	Type FileDiffType
}

// Repository is the API message for repository info.
type Repository struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	FullPath string `json:"fullPath"`
	WebURL   string `json:"webUrl"`
}

// PullRequestFile is the API message for file in the pull request.
type PullRequestFile struct {
	Path         string
	LastCommitID string
	IsDeleted    bool
	// WebURL is the changed file web url in the pull request.
	WebURL string
}

// BranchInfo is the API message for repository branch.
type BranchInfo struct {
	Name         string
	LastCommitID string
}

type PullRequestComment struct {
	ID      string
	Content string
}

type File struct {
	Path    string
	Name    string
	Sha     string
	Content string
}

// Provider is the interface for VCS provider.
type Provider interface {
	// Returns the API URL for a given VCS instance URL
	APIURL(instanceURL string) string

	// Fetch repositories within a given user's scope
	FetchRepositoryList(ctx context.Context, listAll bool) ([]*Repository, error)

	// Reads the file content
	ReadFileContent(ctx context.Context, repositoryID, filePath string, refInfo RefInfo) (string, error)

	// GetBranch gets the given branch in the repository.
	GetBranch(ctx context.Context, repositoryID, branchName string) (*BranchInfo, error)

	// ListPullRequestFile lists changed files in a pull request.
	ListPullRequestFile(ctx context.Context, repositoryID, pullRequestID string) ([]*PullRequestFile, error)

	// CreatePullRequestComment creates a pull request comment.
	CreatePullRequestComment(ctx context.Context, repositoryID, pullRequestID, comment string) error

	// UpdatePullRequestComment updates a pull request comment.
	UpdatePullRequestComment(ctx context.Context, repositoryID, pullRequestID string, comment *PullRequestComment) error

	// ListPullRequestComments lists comments in a pull request.
	ListPullRequestComments(ctx context.Context, repositoryID, pullRequestID string) ([]*PullRequestComment, error)

	// Creates a webhook. Returns the created webhook ID on success.
	CreateWebhook(ctx context.Context, repositoryID string, payload []byte) (string, error)

	// Deletes a webhook.
	DeleteWebhook(ctx context.Context, repositoryID, webhookID string) error
}

var (
	providerMu sync.RWMutex
	providers  = make(map[storepb.VCSType]providerFunc)
)

// ProviderConfig is the provider configuration.
type ProviderConfig struct {
	InstanceURL string
	AuthToken   string
}

type providerFunc func(ProviderConfig) Provider

// Register makes a vcs provider available by the provided type.
// If Register is called twice with the same name or if provider is nil,
// it panics.
func Register(vcsType storepb.VCSType, f providerFunc) {
	providerMu.Lock()
	defer providerMu.Unlock()
	if f == nil {
		panic("vcs: Register provider is nil")
	}
	if _, dup := providers[vcsType]; dup {
		panic(fmt.Sprintf("vcs: Register called twice for provider %q", vcsType.String()))
	}
	providers[vcsType] = f
}

// Get returns a vcs provider specified by its vcs type.
func Get(vcsType storepb.VCSType, providerConfig ProviderConfig) Provider {
	providerMu.RLock()
	f, ok := providers[vcsType]
	providerMu.RUnlock()
	if !ok {
		panic(fmt.Sprintf("vcs: unknown provider %q", vcsType.String()))
	}

	return f(providerConfig)
}
