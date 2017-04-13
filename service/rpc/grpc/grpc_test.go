package grpc_test

import (
	"errors"
	"testing"

	"context"
	"github.com/gkarlik/quark-go"
	rpc "github.com/gkarlik/quark-go/service/rpc/grpc"
	proxy "github.com/gkarlik/quark-go/service/rpc/grpc/test"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type TestRPCService struct {
	*quark.ServiceBase
}

func (s *TestRPCService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	proxy.RegisterTestServiceServer(server.(*grpc.Server), serviceInstance.(proxy.TestServiceServer))

	return nil
}

func (s *TestRPCService) Sum(ctx context.Context, request *proxy.TestRequest) (*proxy.TestResponse, error) {
	return &proxy.TestResponse{
		Sum: request.A + request.B,
	}, nil
}

func TestGRPCServer(t *testing.T) {
	srv := rpc.NewServer()

	addr, err := quark.GetHostAddress(8765)
	assert.NoError(t, err, "Cannot resolve host address")

	ts := &TestRPCService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A"),
			quark.Address(addr)),
	}

	go func() {
		defer srv.Stop()

		conn, err := grpc.Dial(addr.String(), grpc.WithInsecure(), grpc.WithBlock())
		assert.NoError(t, err, "Cannot connect to gRPC server")
		defer conn.Close()

		c := proxy.NewTestServiceClient(conn)
		result, err := c.Sum(context.Background(), &proxy.TestRequest{A: 1, B: 2})

		assert.NoError(t, err, "Error while calling service method")
		assert.Equal(t, int64(3), result.Sum)
	}()

	srv.Start(ts)
}

type InvalidRegisterServiceRPCService struct {
	*quark.ServiceBase
}

func (s *InvalidRegisterServiceRPCService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	return errors.New("Cannot register service")
}

func TestRegisterService(t *testing.T) {
	addr, _ := quark.GetHostAddress(1234)

	srv := rpc.NewServer()
	defer srv.Dispose()

	ws := &InvalidRegisterServiceRPCService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A"),
			quark.Address(addr)),
	}

	assert.Panics(t, func() {
		srv.Start(ws)
	})
}
