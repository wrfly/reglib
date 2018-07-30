package reglib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func parseDockerCondig() (dockerConfig, error) {
	dc := dockerConfig{}

	HOME := os.Getenv("HOME")
	config := path.Join(HOME, ".docker/config.json")
	f, err := os.Open(config)
	if err != nil {
		if os.IsNotExist(err) {
			return dc, nil
		}
		return dc, err
	}
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return dc, err
	}

	if err := json.Unmarshal(bs, &dc); err != nil {
		return dc, err
	}

	return dc, nil
}

func parseAuth(auth string) (string, string) {
	bs, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", ""
	}
	str := fmt.Sprintf("%s", bs)
	uAp := strings.Split(str, ":")
	if len(uAp) != 2 {
		return "", ""
	}

	return uAp[0], uAp[1]
}

// a="1",b="2" -> map["a":"1","b":"2"]
func string2Map(str string) map[string]string {
	pairs := strings.Split(str, ",")
	maps := make(map[string]string, 0)
	for _, pair := range pairs {
		p := strings.Split(pair, "=")
		if len(p) != 2 {
			continue
		}
		k := p[0]
		v := strings.Replace(p[1], "\"", "", -1)
		maps[k] = v
	}

	return maps
}

// GetAuthFromFile returns the username, password of that registry from
// the config file ($HOME/.docker/config.json)
func GetAuthFromFile(regAddr string) (string, string) {
	configs, err := parseDockerCondig()
	if err != nil {
		return "", ""
	}
	for reg, auth := range configs.Auths {
		if reg == regAddr {
			return parseAuth(auth.Auth)
		}
	}
	return "", ""
}

func slice2Chan(slice []string, ch chan string) {
	for _, s := range slice {
		ch <- s
	}
}

func chan2Slice(ch chan string, slice []string, start, end int) {
	i := 0
	for s := range ch {
		if i >= start {
			if i <= end || end < 0 {
				slice = append(slice, s)
			}
		}
		i++
	}
}

func repoChan2Slice(ch chan Repository, slice []Repository, start, end int) {
	i := 0
	for s := range ch {
		if i >= start {
			if i <= end || end < 0 {
				slice = append(slice, s)
			}
		}
		i++
	}
}

// ExtractTagNames returns the names of the []Tags
func ExtractTagNames(tags []Tag) []string {
	names := make([]string, 0, len(tags))
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return names
}
