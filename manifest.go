package reglib

import (
	"context"
	"encoding/json"

	dis "github.com/docker/distribution"
	v1 "github.com/docker/distribution/manifest/schema1"
	v2 "github.com/docker/distribution/manifest/schema2"
)

func manifestV1(ctx context.Context, ms dis.ManifestService,
	tag string) (*v1.Manifest, error) {
	manifestV1 := &v1.Manifest{
		FSLayers: []v1.FSLayer{},
		History:  []v1.History{},
	}
	m, err := ms.Get(ctx, "",
		dis.WithTag(tag),
		dis.WithManifestMediaTypes(
			[]string{v1.MediaTypeManifest},
		))
	if err != nil {
		return nil, err
	}
	_, pld, err := m.Payload()
	if err != nil {
		return nil, err
	}

	return manifestV1, json.Unmarshal(pld, manifestV1)
}

func manifestV2(ctx context.Context, ms dis.ManifestService,
	tag string) (*v2.Manifest, error) {
	manifestV2 := &v2.Manifest{
		Layers: []dis.Descriptor{},
	}
	m, err := ms.Get(ctx, "",
		dis.WithTag(tag),
		dis.WithManifestMediaTypes(
			[]string{v2.MediaTypeManifest},
		))
	if err != nil {
		return nil, err
	}
	_, pld, err := m.Payload()
	if err != nil {
		return nil, err
	}

	return manifestV2, json.Unmarshal(pld, manifestV2)
}
