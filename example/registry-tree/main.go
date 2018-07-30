package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	"github.com/wrfly/reglib"
)

func main() {
	registry := flag.String("r", "r.kfd.me", "registry address")
	user := flag.String("u", "admin", "registry auth username")
	pass := flag.String("p", "admin123", "registry auth password")
	dir := flag.String("d", ".", "file path to store the image info files")
	flag.Parse()

	log.Printf("connect to registry [%s] with [%s:%s] and dumps to dir [%s]\n",
		*registry, *user, *pass, *dir)

	r, err := reglib.New(*registry, *user, *pass)
	if err != nil {
		panic(err)
	}

	repos, err := r.ReposChan(context.Background(), &reglib.ListRepoOptions{
		WithTags: true,
	})
	if err != nil {
		log.Printf("get repos error: %s\n", err)
		return
	}

	base := path.Join(*dir, r.Host())
	for repo := range repos {
		handleRepo(base, repo)
	}

}
func handleRepo(base string, repo reglib.Repository) {
	repoDir := path.Join(base, repo.Name)
	log.Printf("create dir %s\n", repoDir)
	err := os.MkdirAll(repoDir, 0755)
	if err != nil {
		log.Printf("mkdir error: %s\n", err)
		return
	}

	buckets := make(chan struct{}, 100)

	var wg sync.WaitGroup
	wg.Add(len(repo.Tags()))

	for _, tag := range repo.Tags() {
		go func(tag reglib.Tag) {
			buckets <- struct{}{}
			defer func() { <-buckets }()
			defer wg.Done()

			log.Printf("create tag file %s\n", tag.FullName)
			img, err := tag.Image()
			if err != nil {
				log.Printf("get image error: %s\n", err)
				return
			}
			f, err := os.Create(path.Join(repoDir, tag.Name))
			if err != nil {
				log.Printf("create file error: %s\n", err)
				return
			}
			defer f.Close()
			f.WriteString(fmt.Sprintf("fullname: %s\n", tag.FullName))
			f.WriteString(fmt.Sprintf("created: %s\n", img.Created()))
			f.WriteString(fmt.Sprintf("size: %s\n", img.Size()))
		}(tag)
	}

	wg.Wait()
	close(buckets)

}
