package grpc

import (
	"github.com/gkarlik/quark"
	logging "github.com/gkarlik/quark/service/log"
	"google.golang.org/grpc"
	"net"
)

type gRPCServer struct {
}

// NewGRPCServer creates instance of RPC server which is based on gRPC library
func NewGRPCServer() *gRPCServer {
	return &gRPCServer{}
}

func (rpc *gRPCServer) StartRPCService(s quark.RPCService) {
	url, err := s.GetHostAddress()
	if err != nil {
		s.Log().PanicWithFields(logging.LogFields{
			"error": err,
		}, "Cannot resolve service url")
	}

	addr := url.String()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.Log().PanicWithFields(logging.LogFields{
			"error":   err,
			"address": addr,
		}, "Error during listening on port")
	}

	s.Log().Info("Registering gRPC server")

	srv := grpc.NewServer()
	s.RegisterServiceInstance(srv, s)

	s.Log().InfoWithFields(logging.LogFields{"address": addr}, "Listening incomming connections")
	if err := srv.Serve(l); err != nil {
		s.Log().PanicWithFields(logging.LogFields{
			"error": err,
		}, "Failed to serve clients")
	}
}
