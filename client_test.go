package reglib

import (
	"context"
	"os"
	"testing"
)

func TestClient(t *testing.T) {
	r := "docker.nexusguard.net"
	c := client{
		baseURL:  r,
		username: os.Getenv("REG_U"),
		password: os.Getenv("REG_P"),
	}

	if err := c.init(); err != nil {
		t.Errorf("init client error: %s", err)
		return
	}

	ctx := context.Background()

	t.Run("list repos", func(t *testing.T) {
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
	})

	t.Run("list repos with tag", func(t *testing.T) {
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
			for _, tag := range repo.Tags() {
				t.Logf("got [%s:%s]\n", repo.FullName, tag)
			}
		}
	})

}
