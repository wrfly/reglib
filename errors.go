package reglib

import "errors"

// httpsError net/http/client.go +263
var errHTTPS = errors.New("http: server gave HTTP response to HTTPS client")
