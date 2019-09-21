package reglib

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	dis "github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	rClient "github.com/docker/distribution/registry/client"
)

// Client represents for a docker registry client
type Client struct {
	baseURL  string
	username string
	password string

	registry    rClient.Registry
	author      http.RoundTripper
	registryURL *url.URL
	client      *http.Client
}

func (c *Client) init() error {
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

	c.client = &http.Client{
		Transport: c.author,
	}

	return err
}

func (c *Client) Repos(ctx context.Context,
	opts *ListRepoOptions) ([]Repository, error) {

	repoChan, err := c.ReposChan(ctx, opts)
	if err != nil {
		return nil, err
	}
	repos := []Repository{}
	for repo := range repoChan {
		repos = append(repos, repo)
	}
	return repos, nil
}

func (c *Client) ReposChan(ctx context.Context,
	opts *ListRepoOptions) (chan Repository, error) {

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
		size, total = 50, 0
		start, end  = opts.Start, opts.End
		allRepos    = make(chan string, size)
		repoChan    = make(chan Repository)
	)

	go func() {
		defer close(allRepos)
		for {
			tempRepos := make([]string, size)
			n, err := c.registry.Repositories(ctx, tempRepos, last)
			slice2Chan(tempRepos[:n], allRepos)
			if err == io.EOF {
				break
			}
			if err == nil {
				total += n
				if end != 0 && total > end {
					break
				}
				last = tempRepos[n-1]
				continue
			} else {
				log.Printf("get repos error: %s\n", err)
			}
		}
	}()

	go func() {
		var wg sync.WaitGroup
		var i int
		for name := range allRepos {
			if i < start || (i >= end && end > 0) {
				continue
			}
			wg.Add(1)
			go func(repo string) {
				defer wg.Done()
				var tags []Tag
				var err error

				if opts.WithTags {
					tags, err = c.Tags(ctx, repo, nil)
				}

				repoChan <- Repository{
					Namespace: strings.Split(repo, "/")[0],
					Name:      repo,
					tags:      tags,
					cli:       c,
					tagErr:    err,
				}
			}(name)
			i++
		}
		wg.Wait()
		// runtime.Goexit()
		close(repoChan)
	}()

	return repoChan, nil
}

func (c *Client) Tags(ctx context.Context, repo string,
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
		}
		manifestTags = append(manifestTags, Tag{
			FullName: repo + ":" + tag,
			Name:     tag,
			RepoName: repo,
			image:    img,
			cli:      c,
			imgErr:   err,
		})
	}

	return manifestTags, nil

}

func (c *Client) Image(ctx context.Context, repo, tag string) (img *Image, err error) {
	img = &Image{c: c}

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
		return img, fmt.Errorf("get schamev1 error: %s", err)
	}

	img.V2, err = manifestV2(ctx, ms, tag)
	if err != nil {
		return img, fmt.Errorf("get schamev2 error: %s", err)
	}

	return img, err
}

func (c *Client) Host() string {
	return c.registryURL.Host
}

func (c *Client) newRepo(name, tag string) (dis.Repository, error) {
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
