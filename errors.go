package reglib

import "errors"

var (
	errHTTPS  = errors.New("http: server gave HTTP response to HTTPS client")
	errNilCli = errors.New("client is nil")
)
