package grpc

import (
	"net"

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

// Stop stops RPC server.
func (rpc *Server) Stop() {
	if rpc.server != nil {
		rpc.server.Stop()
	}
}

// Start registers and starts service in RPC server.
func (rpc *Server) Start(s quark.RPCService) {
	addr := s.Info().Address.Host

	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"address":   addr,
			"component": componentName,
		}, "Error during listening on port")
	}

	s.Log().InfoWithFields(logger.Fields{"component": componentName}, "Registering gRPC server")

	if err := s.RegisterServiceInstance(rpc.server, s); err != nil {
		s.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Cannot register service instance in RPC server")
	}

	s.Log().InfoWithFields(logger.Fields{
		"address":   addr,
		"component": componentName,
	}, "Listening incomming connections")
	if err := rpc.server.Serve(l); err != nil {
		s.Log().PanicWithFields(logger.Fields{
			"error":     err,
			"component": componentName,
		}, "Failed to serve clients")
	}
}

// Dispose stops server and cleans up RPC server instance
func (rpc *Server) Dispose() {
	logger.Log().InfoWithFields(logger.Fields{"component": componentName}, "Disposing RPC server instance")

	rpc.Stop()

	if rpc.server != nil {
		rpc.server = nil
	}
}
