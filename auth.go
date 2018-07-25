package reglib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

type author struct {
	user, pass string
	client     *http.Client
	tokens     map[string]token
	tokenMutex sync.RWMutex
}

func newAuthRoundTripper(username, password string) *author {
	return &author{
		user:   username,
		pass:   password,
		client: http.DefaultClient,
		tokens: make(map[string]token, 100),
	}
}

func (a *author) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := a.client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "server gave HTTP response to HTTPS client") {
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
	}

	return a.client.Do(req)
}

func (a *author) getAuthString(resp *http.Response) (string, error) {
	challenge := resp.Header.Get("WWW-Authenticate")

	if t := a.checkToken(challenge); t != "" {
		return t, nil
	}

	s := strings.Split(challenge, " ")
	scheme, details := s[0], s[1]
	m := string2Map(details)

	req, _ := http.NewRequest("GET", m["realm"], nil)
	q := req.URL.Query()
	q.Set("service", m["service"])
	q.Set("scope", m["scope"])
	req.URL.RawQuery = q.Encode()

	req.SetBasicAuth(a.user, a.pass)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokenBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	t := token{}
	if err := json.Unmarshal(tokenBytes, &t); err != nil {
		return "", err
	}

	t.scheme = scheme
	a.storeToken(challenge, t)

	return fmt.Sprintf("%s %s", t.scheme, t.Token), nil
}

func (a *author) checkToken(challenge string) string {
	a.tokenMutex.RLock()
	defer a.tokenMutex.RUnlock()
	if t, exist := a.tokens[challenge]; exist {
		return fmt.Sprintf("%s %s", t.scheme, t.Token)
		// the registry's date time may not the same as us, but 300s is sufficient
		// exp := time.Second * time.Duration(t.ExpiresIn)
		// if t.IssuedAt.Sub(time.Now().Add(exp)).Seconds() > 0 {
		// }
	}
	return ""
}

func (a *author) storeToken(challenge string, t token) {
	a.tokenMutex.Lock()
	a.tokens[challenge] = t
	a.tokenMutex.Unlock()
}
