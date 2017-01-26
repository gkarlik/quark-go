package quark

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/gkarlik/quark/metrics"
	"github.com/gkarlik/quark/service/trace"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
)

// GetEnvVar gets environment variable by key. Panics is variable is not set.
func GetEnvVar(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("Environment variable %q is not set!", key))
	}
	return v
}

// GetHostAddress return host and port address on which service is hosted
func GetHostAddress(port int) (*url.URL, error) {
	ip, err := getLocalIPAddress()
	if err != nil {
		return nil, err
	}

	u := fmt.Sprintf("%s:%d", ip, port)
	if port == 0 {
		u = fmt.Sprintf(ip)
	}

	return url.Parse(u)
}

func getLocalIPAddress() (string, error) {
	ifaces, error := net.Interfaces()
	if error != nil {
		return "", error
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, error := iface.Addrs()
		if error != nil {
			return "", error
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Network not available")
}

// ReportServiceValue sends metric with name and value using service Metrics
func ReportServiceValue(s Service, name string, value interface{}) error {
	m := metrics.Metric{
		Name: "response_time",
		Tags: map[string]string{"service": s.Info().Name},
		Values: map[string]interface{}{
			"value": value,
		},
	}
	return s.Metrics().Report(m)
}

// CallHTTPService calls http service at specified url with http method and body
func CallHTTPService(s Service, method string, url string, body io.Reader, parent trace.Span) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	s.Tracer().InjectSpan(parent, opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// StartRPCSpan starts span with name and parent span taken from RPC context
func StartRPCSpan(s Service, name string, ctx context.Context) trace.Span {
	var span trace.Span

	sp := s.Tracer().SpanFromContext(ctx)
	if sp != nil {
		span = s.Tracer().StartSpanWithParent(name, sp)
	} else {
		span = s.Tracer().StartSpan(name)
	}
	return span
}
