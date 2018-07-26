package reglib

import (
	"context"
	"time"

	"github.com/docker/distribution/registry/api/errcode"
)

type Repository struct {
	FullName  string
	Namespace string
	tags      []string
	cli       *client
}

func (r *Repository) Tags() []string {
	if len(r.tags) != 0 {
		return r.tags
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tags, _ := r.cli.Tags(ctx, r.FullName, nil)
	r.tags = tags
	return tags
}

type Image struct {
	Name   string
	Layers []Layer
}

type Layer struct{}

type ListRepoOptions struct {
	WithTags  bool
	Start     int
	End       int
	Namespace string
	Prefix    string
}

type ListTagOptions struct {
	All    bool
	Prefix string
}

type Errors []errcode.Error

type token struct {
	Token     string    `json:"token"`
	ExpiresIn int       `json:"expires_in"`
	IssuedAt  time.Time `json:"issued_at"`
	scheme    string
}

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

type named struct {
	name string
}

func (r *named) String() string {
	return r.name
}

func (r *named) Name() string {
	return r.name
}
