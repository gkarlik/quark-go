package grpc

import (
	"net"
	"strings"

	"github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/logger"
	"google.golang.org/grpc"
)

const componentName = "gRPCServer"

// Server represents RPC server based on gRPC library
type Server struct {
	server *grpc.Server
}

// NewServer creates instance of RPC server which is based on gRPC library.
func NewServer() *Server {
	return &Server{
		server: grpc.NewServer(),
	}
}

// Stop gracefully stops RPC server.
func (rpc *Server) Stop() {
	if rpc.server != nil {
		rpc.server.Stop()
	}
}

// Start registers and starts service in RPC server.
func (rpc *Server) Start(s quark.RPCService) {
	addr := s.Info().Address.String()

	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.Log().PanicWithFields(logger.LogFields{
			"error":     err,
			"address":   addr,
			"component": componentName,
		}, "Error during listening on port")
	}

	s.Log().InfoWithFields(logger.LogFields{"component": componentName}, "Registering gRPC server")

	if err := s.RegisterServiceInstance(rpc.server, s); err != nil {
		s.Log().PanicWithFields(logger.LogFields{
			"error":     err,
			"component": componentName,
		}, "Cannot register service instance in RPC server")
	}

	s.Log().InfoWithFields(logger.LogFields{
		"address":   addr,
		"component": componentName,
	}, "Listening incomming connections")
	// workaround for issue in gRPC library
	if err := rpc.server.Serve(l); !strings.Contains(err.Error(), "use of closed network connection") {
		s.Log().PanicWithFields(logger.LogFields{
			"error":     err,
			"component": componentName,
		}, "Failed to serve clients")
	}
}

// Dispose stops server and cleans up RPC server instance
func (rpc *Server) Dispose() {
	logger.Log().InfoWithFields(logger.LogFields{"component": componentName}, "Disposing RPC server instance")

	rpc.Stop()

	if rpc.server != nil {
		rpc.server = nil
	}
}
