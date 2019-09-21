package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/wrfly/reglib"
)

func watchImage(ctx context.Context, image string) (chan string, error) {
	return nil, nil
}

func main() {
	registry := flag.String("r", "r.kfd.me", "registry address")
	user := flag.String("u", "admin", "registry auth username")
	pass := flag.String("p", "admin123", "registry auth password")
	image := flag.String("image", "alpine:latest", "watch image")
	flag.Parse()

	log.Printf("connect to registry [%s] with [%s:%s]",
		*registry, *user, *pass)

	r, err := reglib.New(*registry, *user, *pass)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(*image, ":") {
		panic("bad image")
	}

	ctx, cancel := context.WithCancel(context.Background())
	// watch changes
	go func() {
		repo := strings.Split(*image, ":")[0]
		tag := strings.Split(*image, ":")[1]
		var created time.Time
		log.Printf("watch [%s:%s] changes", repo, tag)
		for ctx.Err() == nil {
			img, err := r.Image(context.Background(), repo, tag)
			if err != nil {
				panic(err)
			}
			if created.IsZero() {
				created = img.Created()
				log.Printf("image %s created at %s", *image, created)
			}
			if created != img.Created() {
				log.Printf("image %s changed", *image)
				log.Printf("image %s created at %s", *image, created)
				created = img.Created()
			}
			time.Sleep(time.Second)
		}
	}()

	sigC := make(chan os.Signal)
	signal.Notify(sigC, os.Interrupt)

	<-sigC
	cancel()

}
