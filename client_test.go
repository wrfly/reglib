package reglib

import (
	"context"
	"os"
	"testing"
)

func initTestClient() (*Client, error) {
	c := &Client{
		baseURL:  "r.kfd.me",
		username: "admin",
		password: os.Getenv("REG_P"),
	}
	return c, c.init()
}

func TestInitClient(t *testing.T) {
	if _, err := initTestClient(); err != nil {
		t.Error(err)
	}
}

func TestListRepos(t *testing.T) {
	ctx := context.Background()
	c, err := initTestClient()
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("0-10", func(t *testing.T) {
		repos, err := c.Repos(ctx, &ListRepoOptions{
			Start: 0,
			End:   10,
		})
		if err != nil {
			t.Errorf("get repos error: %s", err)
			return
		}
		for _, repo := range repos {
			t.Logf("got [%s]\n", repo.Name)
		}
	})

	t.Run("all", func(t *testing.T) {
		repos, err := c.Repos(ctx, nil)
		if err != nil {
			t.Errorf("get repos error: %s", err)
			return
		}
		for _, repo := range repos {
			t.Logf("got [%s]\n", repo.Name)
		}
	})

}

func TestListRepoWithTags(t *testing.T) {
	ctx := context.Background()
	c, err := initTestClient()
	if err != nil {
		t.Error(err)
		return
	}

	repos, err := c.Repos(ctx, &ListRepoOptions{
		WithTags: true,
		Start:    0,
		End:      10,
	})
	if err != nil {
		t.Errorf("get repos with tags error: %s", err)
		return
	}
	for _, repo := range repos {
		t.Logf("image %s\n", repo.Name)
		tags, err := repo.Tags()
		if err != nil {
			t.Logf("got tag error: %s", err)
			continue
		}
		for _, tag := range tags {
			t.Logf("got tag %s\n", tag.FullName)
		}
	}

}

func TestGetRepoTags(t *testing.T) {
	ctx := context.Background()
	c, err := initTestClient()
	if err != nil {
		t.Error(err)
		return
	}

	tags, err := c.Tags(ctx, "alpine", nil)
	if err != nil {
		t.Logf("got tag error: %s", err)
		return
	}
	for _, tag := range tags {
		t.Logf("got [%s] tag: %s\n", tag.RepoName, tag.Name)
	}

}

func TestGetImage(t *testing.T) {
	ctx := context.Background()
	c, err := initTestClient()
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("client get image", func(t *testing.T) {
		image, err := c.Image(ctx, "alpine", "")
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(image.FullName())
		t.Log(image.Layers())
		for _, hist := range image.History() {
			t.Log(hist.Created)
		}
	})

	t.Run("tag get image", func(t *testing.T) {
		tags, err := c.Tags(ctx, "alpine", nil)
		if err != nil {
			t.Logf("got image error: %s", err)
			return
		}
		tag := tags[0]
		img, err := tag.Image()
		if err != nil {
			t.Logf("got image error: %s", err)
			return
		}
		t.Log(tag.FullName, img.Created())
	})

	t.Run("image size", func(t *testing.T) {
		tags, err := c.Tags(ctx, "alpine", nil)
		if err != nil {
			t.Logf("got image error: %s", err)
			return
		}
		tag := tags[0]
		img, err := tag.Image()
		if err != nil {
			t.Logf("got image error: %s", err)
			return
		}
		t.Logf("size of %s: %d", tag.FullName, img.Size())
		t.Logf("size of %s: %s", tag.FullName, img.Size())
	})
}
