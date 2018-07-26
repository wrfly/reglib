package reglib

import (
	"context"
	"time"

	dis "github.com/docker/distribution"
	v1 "github.com/docker/distribution/manifest/schema1"
	v2 "github.com/docker/distribution/manifest/schema2"
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
	V1 *v1.Manifest
	V2 *v2.Manifest
}

func (i *Image) FullName() string {
	return i.V1.Name + i.V1.Tag
}

func (i *Image) History() []v1.History {
	return i.V1.History
}

func (i *Image) FSLayers() []v1.FSLayer {
	return i.V1.FSLayers
}

func (i *Image) Layers() []dis.Descriptor {
	return i.V2.Layers
}

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
