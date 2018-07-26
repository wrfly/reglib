package reglib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	rClient "github.com/docker/distribution/registry/client"
	"github.com/opencontainers/go-digest"
)

type client struct {
	baseURL  string
	username string
	password string

	registry    rClient.Registry
	author      http.RoundTripper
	registryURL *url.URL
}

func (c *client) init() error {
	if c.username == "" || c.password == "" {
		u, p, err := GetAuthFromFile(c.baseURL)
		if err != nil {
			return err
		}
		c.username, c.password = u, p
	}
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	if u.Scheme == "" {
		c.baseURL = "https://" + c.baseURL
	}

	c.registryURL, _ = url.Parse(c.baseURL)

	c.author = newAuthRoundTripper(c.username, c.password)
	c.registry, err = rClient.NewRegistry(c.baseURL, c.author)

	slice := make([]string, 1)
	n, err := c.registry.Repositories(context.Background(), slice, "")
	if n != 1 {
		return fmt.Errorf("can not get repositories")
	}
	return err
}

func (c *client) Repos(ctx context.Context, opts *ListRepoOptions) ([]Repository, error) {
	if opts == nil {
		opts = &ListRepoOptions{}
	} else {
		// check opts
		if opts.Start > opts.End {
			return nil, fmt.Errorf("invalid start(%d) and end(%d)", opts.Start, opts.End)
		}
	}

	var (
		last        = ""
		total       = 0
		allRepos    = []string{}
		targetRepos = []string{}
		repos       = []Repository{}
	)

	for {
		tempRepos := make([]string, 10)
		n, err := c.registry.Repositories(ctx, tempRepos, last)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		allRepos = append(allRepos, tempRepos...)
		total += n
		if opts.End != 0 && total > opts.End {
			break
		}
		last = tempRepos[n-1]
	}

	if opts.Start < opts.End {
		targetRepos = allRepos[opts.Start:opts.End]
	} else {
		targetRepos = allRepos
	}

	var wg sync.WaitGroup
	wg.Add(len(targetRepos))
	repoChan := make(chan Repository)

	for _, name := range targetRepos {
		go func(repo string) {
			if opts.WithTags {
				tags, _ := c.Tags(ctx, repo, nil)
				repoChan <- Repository{
					FullName: repo,
					tags:     tags,
					cli:      c,
				}
			} else {
				repoChan <- Repository{FullName: repo, cli: c}
			}
			wg.Done()
		}(name)
	}

	consumeChan := make(chan struct{})
	go func() {
		for repo := range repoChan {
			repos = append(repos, repo)
		}
		consumeChan <- struct{}{}
	}()

	wg.Wait()
	close(repoChan)

	<-consumeChan
	return repos, nil
}

func (c *client) Tags(ctx context.Context, repository string, opts *ListTagOptions) ([]string, error) {
	r, err := rClient.NewRepository(&named{name: repository}, c.baseURL, c.author)
	if err != nil {
		return nil, err
	}
	return r.Tags(ctx).All(ctx)
}

func (c *client) Image(ctx context.Context, repository, tag string) (img Image, err error) {
	r, err := rClient.NewRepository(&named{name: repository}, c.baseURL, c.author)
	if err != nil {
		return
	}
	m, err := r.Manifests(ctx)
	if err != nil {
		return
	}
	manifest, err := m.Get(ctx, digest.FromString(tag))
	if err != nil {
		return
	}
	descriptors := manifest.References()

	for _, des := range descriptors {
		fmt.Printf("%v\n", des)
	}

	return
}

func (c *client) RegistryAddress() string {
	return c.registryURL.Host
}
