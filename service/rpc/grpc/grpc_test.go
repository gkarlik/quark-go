package grpc_test

import (
	"errors"
	"github.com/gkarlik/quark"
	"github.com/gkarlik/quark/logger/logrus"
	"github.com/gkarlik/quark/service"
	rpc "github.com/gkarlik/quark/service/rpc/grpc"
	proxy "github.com/gkarlik/quark/service/rpc/grpc/test"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"testing"
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

	ts := &TestRPCService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A"),
			quark.Port(8765),
			quark.Logger(logrus.NewLogger())),
	}

	addr, err := ts.GetHostAddress()
	assert.NoError(t, err, "Cannot resolve host address")

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

type WrongRegisterServiceRPCService struct {
	*quark.ServiceBase
}

func (s *WrongRegisterServiceRPCService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	return errors.New("Cannot register service")
}

func TestRegisterService(t *testing.T) {
	srv := rpc.NewServer()
	defer srv.Stop()

	ws := &WrongRegisterServiceRPCService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A"),
			quark.Port(8765),
			quark.Logger(logrus.NewLogger())),
	}

	assert.Panics(t, func() {
		srv.Start(ws)
	})
}

type WrongHostAddressRPCService struct {
	*quark.ServiceBase
}

func (s *WrongHostAddressRPCService) GetHostAddress() (service.Address, error) {
	return nil, errors.New("Cannot resolve host address")
}

func (s *WrongHostAddressRPCService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	return nil
}

func TestGetHostAddress(t *testing.T) {
	srv := rpc.NewServer()
	defer srv.Stop()

	ws := &WrongHostAddressRPCService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Tags("A"),
			quark.Port(8765),
			quark.Logger(logrus.NewLogger())),
	}

	assert.Panics(t, func() {
		srv.Start(ws)
	})
}
