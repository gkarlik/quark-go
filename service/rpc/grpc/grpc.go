package grpc

import (
	"github.com/gkarlik/quark"
	logging "github.com/gkarlik/quark/logger"
	"google.golang.org/grpc"
	"net"
	"strings"
)

// Server represents RPC server based on gRPC library
type Server struct {
	server *grpc.Server
}

// NewServer creates instance of RPC server which is based on gRPC library
func NewServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}

// Stop gracefully stops RPC server
func (rpc *Server) Stop() {
	if rpc.server != nil {
		rpc.server.Stop()
	}
}

// Start registers and starts service in RPC server
func (rpc *Server) Start(s quark.RPCService) {
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

	if err := s.RegisterServiceInstance(rpc.server, s); err != nil {
		s.Log().PanicWithFields(logging.LogFields{"error": err}, "Cannot register service instance in RPC server")
	}

	s.Log().InfoWithFields(logging.LogFields{"address": addr}, "Listening incomming connections")
	// workaround for issue in gRPC library
	if err := rpc.server.Serve(l); !strings.Contains(err.Error(), "use of closed network connection") {
		s.Log().PanicWithFields(logging.LogFields{"error": err}, "Failed to serve clients")
	}
}
