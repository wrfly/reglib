package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/wrfly/reglib"
)

func main() {
	registry := flag.String("r", "r.kfd.me", "registry address")
	user := flag.String("u", "admin", "registry auth username")
	pass := flag.String("p", "admin123", "registry auth password")
	target := flag.String("t", "alpine:latest", "target image")
	flag.Parse()

	targetTag := "latest"
	x := strings.Split(*target, ":")
	targetRepo := x[0]
	if len(x) == 2 {
		targetTag = x[1]
	}

	log.Printf("connect to registry [%s] with [%s:%s]\n",
		*registry, *user, *pass)

	r, err := reglib.New(*registry, *user, *pass)
	if err != nil {
		panic(err)
	}

	image, err := r.Image(context.Background(), targetRepo, targetTag)
	if err != nil {
		panic(err)
	}

	err = image.Download(context.Background(),
		fmt.Sprintf("/tmp/reglib/%s",
			strings.Replace(targetRepo, "/", ".", -1)),
	)
	if err != nil {
		panic(err)
	}

}
