package grpc

import (
	"encoding/base64"
	"github.com/gkarlik/quark"
	logging "github.com/gkarlik/quark/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

type MetadataReaderWriter struct {
	MD *metadata.MD
}

func (w MetadataReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	if strings.HasSuffix(key, "-bin") {
		val = string(base64.StdEncoding.EncodeToString([]byte(val)))
	}

	(*w.MD)[key] = append((*w.MD)[key], val)
}

func (w MetadataReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range *w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
