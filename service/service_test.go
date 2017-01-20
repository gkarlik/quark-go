package service_test

import (
	"github.com/gkarlik/quark/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServiceAddress(t *testing.T) {
	var cases = []struct {
		in, want string
	}{
		{"", ""},
		{"a", "a"},
		{"http://test.url/", "http://test.url/"},
	}

	for _, c := range cases {
		addr := service.NewURIServiceAddress(c.in)
		got := addr.String()

		assert.Equal(t, c.want, got)
	}
}
