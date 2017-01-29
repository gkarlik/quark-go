package service

import (
	"net/url"
)

// Info represents information about service.
type Info struct {
	Name    string   // service name
	Version string   // service version
	Tags    []string // service tags
	Address *url.URL // service address
}
