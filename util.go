package reglib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
)

const dockerConfigPath = ".docker/config.json"

func parseDockerConfig() (dockerConfig, error) {
	dc := dockerConfig{}

	config := path.Join(os.Getenv("HOME"), dockerConfigPath)
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

	// remove prefix and sufix
	for addr, auth := range dc.Auths {
		u, err := url.Parse(addr)
		if err != nil {
			continue
		}

		if u.Scheme != "" {
			if u.Port() == "" {
				addr = u.Hostname()
			} else {
				addr = fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
			}
		}

		// special for docker.io
		if addr == "index.docker.io" {
			dc.Auths["docker.io"] = auth
		}

		dc.Auths[addr] = auth
	}

	return dc, nil
}

// UnmarshalAuth returns the username and password for that auth credential
func UnmarshalAuth(auth string) (string, string) {
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
func GetAuthFromFile(registry string) (string, string) {
	tokens, err := GetAuthTokens()
	if err != nil {
		return "", ""
	}
	return UnmarshalAuth(tokens[registry])
}

// GetAuthTokens from $HOME/.docker/config.json
func GetAuthTokens() (map[string]string, error) {
	cfg, err := parseDockerConfig()
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for addr, auth := range cfg.Auths {
		m[addr] = auth.Auth
	}
	return m, nil
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

// DEBUG enables debug output
var DEBUG bool

func debug(format string, v ...interface{}) {
	if DEBUG {
		log.Printf(format, v...)
	}
}
