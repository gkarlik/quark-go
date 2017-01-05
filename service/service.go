package service

// Disposer reporesents object cleanup mechanism
type Disposer interface {
	Dispose()
}

// Info represents information about service
type Info struct {
	Name    string
	Version string
	Tags    []string
	Port    int
}

// Address represents uri of the service
type Address interface {
	String() string
}

type uriServiceAddress struct {
	uri string
}

func (a *uriServiceAddress) String() string {
	return a.uri
}

// NewURIServiceAddress creates instance of Address based on uri
func NewURIServiceAddress(uri string) *uriServiceAddress {
	return &uriServiceAddress{
		uri: uri,
	}
}
