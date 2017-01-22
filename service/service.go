package service

import (
	"net/url"
)

// Info represents information about service
type Info struct {
	Name    string
	Version string
	Tags    []string
	Address *url.URL
}
