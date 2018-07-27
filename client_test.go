package reglib

import (
	"context"
	"os"
	"testing"
)

func initTestClient() (client, error) {
	r := "docker.nexusguard.net"
	c := client{
		baseURL:  r,
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

	repos, err := c.Repos(ctx, &ListRepoOptions{
		Start: 0,
		End:   10,
	})
	if err != nil {
		t.Errorf("get repos error: %s", err)
		return
	}
	for _, repo := range repos {
		t.Logf("got [%s]\n", repo.FullName)
	}

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
		t.Logf("image %s\n", repo.FullName)
		for _, tag := range repo.Tags() {
			t.Logf("got tag %s\n", tag.FullName)
		}
	}

}

func TestGetImage(t *testing.T) {
	ctx := context.Background()
	c, err := initTestClient()
	if err != nil {
		t.Error(err)
		return
	}

	image, err := c.Image(ctx, "admin/alpine", "")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(image.FullName())
	t.Log(image.Layers())
	for _, hist := range image.History() {
		t.Log(hist.Created)
	}
}
