package reglib

import (
	"context"
	"os"
	"testing"
)

func initTestClient() (client, error) {
	c := client{
		baseURL:  "docker.nexusguard.net",
		username: os.Getenv("REG_U"),
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
		for _, tag := range repo.Tags() {
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

	tags, err := c.Tags(ctx, "platform/qimen", nil)
	if err != nil {
		t.Error(err)
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
		image, err := c.Image(ctx, "platform/crek", "")
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
		tags, err := c.Tags(ctx, "admin/alpine", nil)
		if err != nil {
			t.Error(err)
			return
		}
		tag := tags[0]
		img, err := tag.Image()
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(tag.FullName, img.Created())
	})
}
