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

func parseAuth(auth string) (string, string, error) {
	bs, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		return "", "", err
	}
	str := fmt.Sprintf("%s", bs)
	uAp := strings.Split(str, ":")
	if len(uAp) != 2 {
		return "", "", nil
	}

	return uAp[0], uAp[1], nil
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

func GetAuthFromFile(regAddr string) (string, string, error) {
	configs, err := parseDockerCondig()
	if err != nil {
		return "", "", err
	}
	for reg, auth := range configs.Auths {
		if reg == regAddr {
			return parseAuth(auth.Auth)
		}
	}
	return "", "", fmt.Errorf("not found")
}
