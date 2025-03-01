package lsp

import (
	"sync/atomic"

	"github.com/khulnasoft/devsecdb/backend/component/config"
	"github.com/khulnasoft/devsecdb/backend/store"
)

// Server is the Language Server Protocol service.
type Server struct {
	connectionCount atomic.Uint64

	store   *store.Store
	profile *config.Profile
}

// NewServer creates a Language Server Protocol service.
func NewServer(
	store *store.Store,
	profile *config.Profile,
) *Server {
	return &Server{
		store:   store,
		profile: profile,
	}
}
