package reglib

import (
	"context"
	"fmt"
)

// Registry is the interface of registry client
type Registry interface {
	// Repos list the repositories
	Repos(ctx context.Context, opts *ListRepoOptions) ([]Repository, error)
	// Tags list the tags of the repository
	Tags(ctx context.Context, repository string, opts *ListTagOptions) ([]string, error)
}

// New docker registry client
func New(base, user, pass string) (Registry, error) {
	c := &client{
		baseURL:  base,
		username: user,
		password: pass,
	}

	if err := c.init(); err != nil {
		return nil, fmt.Errorf("init client error: %s", err)
	}

	return c, nil
}
