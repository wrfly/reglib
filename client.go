package reglib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	dis "github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	rClient "github.com/docker/distribution/registry/client"
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
		c.username, c.password = GetAuthFromFile(c.baseURL)
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
	if err == io.EOF {
		return nil
	}
	if n != 1 {
		return fmt.Errorf("can not get repositories: %s (%d)", err, n)
	}
	return err
}

func (c *client) Repos(ctx context.Context,
	opts *ListRepoOptions) ([]Repository, error) {

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
		tempRepos := make([]string, 50)
		n, err := c.registry.Repositories(ctx, tempRepos, last)
		allRepos = append(allRepos, tempRepos[:n]...)
		if err == io.EOF {
			break
		}
		if err == nil {
			total += n
			if opts.End != 0 && total > opts.End {
				break
			}
			last = tempRepos[n-1]
			continue
		} else {
			return nil, err
		}
	}

	if opts.Start < opts.End {
		var start, end, l = opts.Start, opts.End, len(allRepos)
		if l < end {
			end = l
		}
		if start > l {
			start = 0
		}
		targetRepos = allRepos[start:end]
	} else {
		targetRepos = allRepos
	}

	var wg sync.WaitGroup
	wg.Add(len(targetRepos))
	repoChan := make(chan Repository)

	for _, name := range targetRepos {
		go func(repo string) {
			if opts.WithTags {
				tags, err := c.Tags(ctx, repo, nil)
				if err != nil {
					fmt.Println("get tag error:", err)
				}
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

func (c *client) Tags(ctx context.Context, repo string,
	opts *ListTagOptions) ([]Tag, error) {

	if opts == nil {
		opts = &ListTagOptions{}
	}

	named, err := reference.WithName(repo)
	if err != nil {
		return nil, err
	}

	r, err := rClient.NewRepository(named, c.baseURL, c.author)
	if err != nil {
		return nil, err
	}

	tags, err := r.Tags(ctx).All(ctx)
	if err != nil {
		return nil, err
	}

	manifestTags := make([]Tag, 0, len(tags))
	var img *Image
	for _, tag := range tags {
		if opts.WithManifest {
			img, err = c.Image(ctx, repo, tag)
			if err != nil {
				fmt.Printf("get image [%s:%s] error: %s", repo, tag, err)
			}
		}
		manifestTags = append(manifestTags, Tag{
			FullName: repo + ":" + tag,
			TagName:  tag,
			RepoName: repo,
			Image:    img,
		})
	}

	return manifestTags, nil

}

func (c *client) Image(ctx context.Context, repo, tag string) (img *Image, err error) {
	img = &Image{}

	if tag == "" {
		tag = "latest"
	}
	r, err := c.newRepo(repo, tag)
	if err != nil {
		return img, err
	}

	ms, err := r.Manifests(ctx)
	if err != nil {
		return img, err
	}

	img.V1, err = manifestV1(ctx, ms, tag)
	if err != nil {
		return img, err
	}

	img.V2, err = manifestV2(ctx, ms, tag)
	if err != nil {
		return img, err
	}

	return img, err
}

func (c *client) Host() string {
	return c.registryURL.Host
}

func (c *client) newRepo(name, tag string) (dis.Repository, error) {
	named, err := reference.WithName(name)
	if err != nil {
		return nil, err
	}
	nt, err := reference.WithTag(named, tag)
	if err != nil {
		return nil, err
	}
	return rClient.NewRepository(nt, c.baseURL, c.author)
}
