package reglib // import "github.com/wrfly/reglib"

import (
	"context"
	"fmt"
)

// Registry is the interface of registry client
type Registry interface {
	// Repos list the repositories
	Repos(ctx context.Context, opts *ListRepoOptions) ([]Repository, error)
	// ReposChan returns a channel caontains the repos
	ReposChan(ctx context.Context, opts *ListRepoOptions) (chan Repository, error)
	// Tags list the tags of the repository
	Tags(ctx context.Context, repo string, opts *ListTagOptions) ([]Tag, error)
	// Image get the image instance via the specfic repo and tag
	Image(ctx context.Context, repo, tag string) (*Image, error)
	// return the registry's host (domain)
	Host() string
}

// New docker registry client
func New(baseURL, user, pass string) (Registry, error) {
	c := &client{
		baseURL:  baseURL,
		username: user,
		password: pass,
	}

	if err := c.init(); err != nil {
		return nil, fmt.Errorf("init client error: %s", err)
	}

	return c, nil
}
