package grpc

import (
	"net"
	"strings"

	"github.com/gkarlik/quark-go"
	logging "github.com/gkarlik/quark-go/logger"
	"google.golang.org/grpc"
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
	addr := s.Info().Address.String()

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
