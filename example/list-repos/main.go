package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/wrfly/reglib"
)

func main() {
	registry := flag.String("r", "r.kfd.me", "registry address")
	user := flag.String("u", "admin", "registry auth username")
	pass := flag.String("p", "admin123", "registry auth password")
	flag.Parse()

	log.Printf("connect to registry [%s] with [%s:%s]\n",
		*registry, *user, *pass)

	r, err := reglib.New(*registry, *user, *pass)
	if err != nil {
		panic(err)
	}

	repos, err := r.ReposChan(context.Background(), &reglib.ListRepoOptions{
		WithTags: true,
	})
	if err != nil {
		return
	}
	log.Printf("get repos info chan...\n\n")

	for repo := range repos {
		if tags := repo.Tags(); len(tags) > 5 {
			fmt.Printf("%s %v... and %d more\n", repo.Name, reglib.ExtractTagNames(tags[:5]), len(tags)-5)
		} else {
			fmt.Printf("%s %v\n", repo.Name, reglib.ExtractTagNames(tags))
		}
	}

}
