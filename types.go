package reglib

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	dis "github.com/docker/distribution"
	v1 "github.com/docker/distribution/manifest/schema1"
	v2 "github.com/docker/distribution/manifest/schema2"
)

// Repository is the instance of an repo
type Repository struct {
	Name      string
	Namespace string
	tags      []Tag
	cli       *client
}

// Tags returns the repo's tags
func (r *Repository) Tags() []Tag {
	if len(r.tags) != 0 {
		return r.tags
	}
	if r.cli == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	tags, _ := r.cli.Tags(ctx, r.Name, nil)
	r.tags = tags
	return tags
}

// Tag is the image's specfic tag
type Tag struct {
	Name     string
	FullName string
	RepoName string
	image    *Image
	cli      *client
}

// Image returns the repo:tag's manifest
func (t *Tag) Image() (*Image, error) {
	if t.image != nil {
		return t.image, nil
	}
	if t.cli == nil {
		return nil, errNilCli
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	img, err := t.cli.Image(ctx, t.RepoName, t.Name)
	t.image = img
	return img, err
}

type imageSize int64

func (is imageSize) String() string {
	switch {
	case is > gbSize:
		return fmt.Sprintf("%.3fGB", float64(is/gbSize)+float64(is%gbSize)/float64(gbSize))
	case is > mbSize:
		return fmt.Sprintf("%.3fMB", float64(is/mbSize)+float64(is%mbSize)/float64(mbSize))
	case is > kbSize:
		return fmt.Sprintf("%.3fKB", float64(is/kbSize)+float64(is%kbSize)/float64(kbSize))
	default:
		return fmt.Sprintf("%dBytes", is)
	}
}

// Image instance, includes the schemav1 and schemav2
type Image struct {
	V1      *v1.Manifest
	V2      *v2.Manifest
	history []ImageHistory
	size    imageSize
}

// FullName return the image name and it's tag
func (i *Image) FullName() string {
	if i.V1 == nil {
		return "error: cannot get name"
	}
	return i.V1.Name + ":" + i.V1.Tag
}

// History converts the v1's history info to reglib's history struct
func (i *Image) History() []ImageHistory {
	if i.V1 == nil {
		return nil
	}
	if len(i.history) != 0 {
		return i.history
	}
	iHistory := make([]ImageHistory, 0)
	for _, hist := range i.V1.History {
		ihist := ImageHistory{}
		if json.Unmarshal([]byte(hist.V1Compatibility), &ihist) == nil {
			iHistory = append(iHistory, ihist)
		}
	}
	return iHistory
}

// FSLayers returns the fs layer info (schemav1)
func (i *Image) FSLayers() []v1.FSLayer {
	if i.V1 == nil {
		return nil
	}
	return i.V1.FSLayers
}

// Layers returns the layer info (schemav2)
func (i *Image) Layers() []dis.Descriptor {
	if i.V2 == nil {
		return nil
	}
	return i.V2.Layers
}

// Created returns the image's create time
func (i *Image) Created() time.Time {
	if i.V1 == nil {
		return time.Time{}
	}
	hist := i.History()
	return hist[len(hist)-1].Created
}

// Size returns the image's size
func (i *Image) Size() imageSize {
	if i.V2 == nil {
		return 0
	}
	if i.size != 0 {
		return i.size
	}
	var size int64
	for _, layer := range i.V2.Layers {
		size += layer.Size
	}
	i.size = imageSize(size)
	return i.size
}

// ListRepoOptions ...
type ListRepoOptions struct {
	WithTags   bool
	Start, End int
	Namespace  string
	Prefix     string
}

// ListTagOptions ...
type ListTagOptions struct {
	WithManifest bool
	Prefix       string
}

type token struct {
	Token     string    `json:"token,omitempty"`
	ExpiresIn int       `json:"expires_in,omitempty"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
	Error     string    `json:"error,omitempty"`
	typ       string
}

type dockerConfig struct {
	Auths map[string]struct {
		Auth string `json:"auth"`
	} `json:"auths"`
}

// ImageHistory converted from json, thanks to https://mholt.github.io/json-to-go/
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
