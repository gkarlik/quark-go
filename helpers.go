package quark

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/service/trace"
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

// GetHostAddress returns host address and optionally port on which service is hosted.
// If port is 0 only address is returned.
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

// ReportServiceValue sends metric with name and value using service instance.
func ReportServiceValue(s Service, name string, value interface{}) error {
	m := metrics.Metric{
		Name: name,
		Tags: map[string]string{"service": s.Info().Name},
		Values: map[string]interface{}{
			"value": value,
		},
	}
	return s.Metrics().Report(m)
}

// CallHTTPService calls HTTP service at specified url with HTTP method and body.
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

// RPCMetadataCarrier represents carrier for span propagation using gRPC metadata.
type RPCMetadataCarrier struct {
	MD *metadata.MD // gRPC metadata
}

// Set sets metadata value inside gRPC metadata.
func (c RPCMetadataCarrier) Set(key, val string) {
	k := strings.ToLower(key)
	if strings.HasSuffix(k, "-bin") {
		val = string(base64.StdEncoding.EncodeToString([]byte(val)))
	}

	(*c.MD)[k] = append((*c.MD)[k], val)
}

// ForeachKey iterates over gRPC metadata key and values.
func (c RPCMetadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range *c.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// StartRPCSpan starts span with name and parent span taken from RPC context.
func StartRPCSpan(s Service, name string, ctx context.Context) trace.Span {
	var span trace.Span
	var err error

	md, ok := metadata.FromContext(ctx)
	if ok {
		span, err = s.Tracer().ExtractSpan(name, opentracing.TextMap, RPCMetadataCarrier{MD: &md})
	}

	if err != nil || !ok {
		span = s.Tracer().StartSpan(name)
	}

	return span
}
