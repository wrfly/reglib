package reglib

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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

// GetAuthFromFile returns the username, password of that registry from
// the config file ($HOME/.docker/config.json)
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

// checkHTTPRedirect is a callback that can manipulate redirected HTTP
// requests. It is used to preserve Accept and Range headers.
func checkHTTPRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}

	if len(via) > 0 {
		for headerName, headerVals := range via[0].Header {
			if headerName != "Accept" && headerName != "Range" {
				continue
			}
			for _, val := range headerVals {
				// Don't add to redirected request if redirected
				// request already has a header with the same
				// name and value.
				hasValue := false
				for _, existingVal := range req.Header[headerName] {
					if existingVal == val {
						hasValue = true
						break
					}
				}
				if !hasValue {
					req.Header.Add(headerName, val)
				}
			}
		}
	}

	return nil
}
