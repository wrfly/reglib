package reglib

import (
	"time"

	"github.com/docker/distribution/registry/api/errcode"
)

type Repository struct {
	FullName  string
	Namespace string
	Tags      []string
}

func (r *Repository) String() string {
	return r.FullName
}

func (r *Repository) Name() string {
	return r.FullName
}

type ListRepoOptions struct {
	NotAll    bool
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
	Token      string    `json:"token"`
	ExpiresIn  int       `json:"expires_in"`
	IssuedAt   time.Time `json:"issued_at"`
	authString string
}

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}
