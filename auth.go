package reglib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type author struct {
	userInfo   *url.Userinfo
	client     *http.Client
	tokens     map[string]token
	tokenMutex sync.RWMutex
}

func newAuthRoundTripper(u, p string) *author {
	return &author{
		userInfo: url.UserPassword(u, p),
		client:   http.DefaultClient,
		tokens:   make(map[string]token, 100),
	}
}

func (a *author) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.User = a.userInfo
	resp, err := a.client.Do(req)
	if err != nil {
		if err == errHTTPS {
			req.URL.Scheme = "http"
			resp, err = a.client.Do(req)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		authString, err := a.getAuthString(resp)
		if err != nil {
			return resp, err
		}
		req.Header.Set("Authorization", authString)
		return a.client.Do(req)
	}

	return resp, nil
}

func (a *author) getAuthString(resp *http.Response) (string, error) {
	challenge := resp.Header.Get("WWW-Authenticate")

	if t := a.checkToken(challenge); t != "" {
		return t, nil
	}

	s := strings.Split(challenge, " ")
	authType, details := s[0], s[1]
	m := string2Map(details)

	if m["realm"] == registryRealm {
		// we have already set the userinfo, just return an empty
		// auth token and the client will add the auth header by default
		return "", nil
	}

	req, err := http.NewRequest("GET", m["realm"], nil)
	q := req.URL.Query()
	q.Set("service", m["service"])
	q.Set("scope", m["scope"])
	req.URL.RawQuery = q.Encode()
	req.URL.User = a.userInfo

	authResp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer authResp.Body.Close()

	tokenBytes, err := ioutil.ReadAll(authResp.Body)
	if err != nil {
		return "", err
	}
	t := token{}
	if err := json.Unmarshal(tokenBytes, &t); err != nil {
		return "", err
	}
	if t.Error != "" {
		fmt.Printf("auth error: %s\n", t.Error)
	}
	// modify the issue_at time, since the data time of the registry may not
	// the same as ours
	t.IssuedAt = time.Now()

	t.typ = authType
	a.storeToken(challenge, t)

	return fmt.Sprintf("%s %s", t.typ, t.Token), nil
}

func (a *author) checkToken(challenge string) string {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()
	if t, exist := a.tokens[challenge]; exist {
		exp := time.Second * time.Duration(t.ExpiresIn)
		if t.IssuedAt.Sub(time.Now().Add(exp)).Seconds() > 0 {
			return fmt.Sprintf("%s %s", t.typ, t.Token)
		}
	}
	return ""
}

func (a *author) storeToken(challenge string, t token) {
	a.tokenMutex.Lock()
	a.tokens[challenge] = t
	a.tokenMutex.Unlock()
}
