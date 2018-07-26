package reglib

import (
	"context"
	"encoding/json"
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
	V1      *v1.Manifest
	V2      *v2.Manifest
	history []ImageHistory
}

func (i *Image) FullName() string {
	return i.V1.Name + i.V1.Tag
}

func (i *Image) History() []ImageHistory {
	if len(i.history) != 0 {
		return i.history
	}
	iHistory := make([]ImageHistory, 0, len(i.V1.History))
	for _, hist := range i.V1.History {
		ihist := ImageHistory{}
		if json.Unmarshal([]byte(hist.V1Compatibility), &ihist) == nil {
			iHistory = append(iHistory, ihist)
		}
	}
	return iHistory
}

func (i *Image) FSLayers() []v1.FSLayer {
	return i.V1.FSLayers
}

func (i *Image) Layers() []dis.Descriptor {
	return i.V2.Layers
}

// func (i *Image) Created() time.Time {
// 	return i.V1.
// }

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

// thanks to https://mholt.github.io/json-to-go/
type ImageHistory struct {
	Architecture string `json:"architecture,omitempty"`
	Config       struct {
		Hostname     string      `json:"Hostname,omitempty"`
		Domainname   string      `json:"Domainname,omitempty"`
		User         string      `json:"User,omitempty"`
		AttachStdin  bool        `json:"AttachStdin,omitempty"`
		AttachStdout bool        `json:"AttachStdout,omitempty"`
		AttachStderr bool        `json:"AttachStderr,omitempty"`
		Tty          bool        `json:"Tty,omitempty"`
		OpenStdin    bool        `json:"OpenStdin,omitempty"`
		StdinOnce    bool        `json:"StdinOnce,omitempty"`
		Env          []string    `json:"Env,omitempty"`
		Cmd          []string    `json:"Cmd,omitempty"`
		ArgsEscaped  bool        `json:"ArgsEscaped,omitempty"`
		Image        string      `json:"Image,omitempty"`
		Volumes      interface{} `json:"Volumes,omitempty"`
		WorkingDir   string      `json:"WorkingDir,omitempty"`
		Entrypoint   interface{} `json:"Entrypoint,omitempty"`
		OnBuild      interface{} `json:"OnBuild,omitempty"`
		Labels       interface{} `json:"Labels,omitempty"`
	} `json:"config,omitempty"`
	Container       string `json:"container,omitempty"`
	ContainerConfig struct {
		Hostname     string      `json:"Hostname,omitempty"`
		Domainname   string      `json:"Domainname,omitempty"`
		User         string      `json:"User,omitempty"`
		AttachStdin  bool        `json:"AttachStdin,omitempty"`
		AttachStdout bool        `json:"AttachStdout,omitempty"`
		AttachStderr bool        `json:"AttachStderr,omitempty"`
		Tty          bool        `json:"Tty,omitempty"`
		OpenStdin    bool        `json:"OpenStdin,omitempty"`
		StdinOnce    bool        `json:"StdinOnce,omitempty"`
		Env          []string    `json:"Env,omitempty"`
		Cmd          []string    `json:"Cmd,omitempty"`
		ArgsEscaped  bool        `json:"ArgsEscaped,omitempty"`
		Image        string      `json:"Image,omitempty"`
		Volumes      interface{} `json:"Volumes,omitempty"`
		WorkingDir   string      `json:"WorkingDir,omitempty"`
		Entrypoint   interface{} `json:"Entrypoint,omitempty"`
		OnBuild      interface{} `json:"OnBuild,omitempty"`
		Labels       struct {
		} `json:"Labels,omitempty"`
	} `json:"container_config,omitempty"`
	Created       time.Time `json:"created,omitempty"`
	DockerVersion string    `json:"docker_version,omitempty"`
	ID            string    `json:"id,omitempty"`
	Os            string    `json:"os,omitempty"`
	Parent        string    `json:"parent,omitempty"`
	Throwaway     bool      `json:"throwaway,omitempty"`
}
