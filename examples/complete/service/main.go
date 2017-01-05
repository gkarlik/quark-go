package main

import (
	"github.com/gkarlik/quark"
	proxy "github.com/gkarlik/quark/examples/complete/service/definitions/proxies/sum"
	"github.com/gkarlik/quark/service/bus/rabbitmq"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/gkarlik/quark/service/discovery/consul"
	"github.com/gkarlik/quark/service/log"
	"github.com/gkarlik/quark/service/log/logrus"
	"github.com/gkarlik/quark/service/metrics/influxdb"
	gRPC "github.com/gkarlik/quark/service/rpc/grpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"os"
	"strconv"
)

type SumService struct {
	*quark.ServiceBase
}

func NewSumService() *SumService {
	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	return &SumService{
		ServiceBase: quark.NewService(
			quark.Name(os.Getenv("SERVICE_NAME")),
			quark.Version(os.Getenv("SERVICE_VERSION")),
			quark.Port(port),
			quark.Logger(logrus.NewLogger()),
			quark.Discovery(consul.NewServiceDiscovery(os.Getenv("CONSUL_ADDRESS"))),
			quark.Bus(rabbitmq.NewMessageBus(os.Getenv("RABBITMQ_ADDRESS"))),
			quark.Metrics(influxdb.NewMetricsReporter(os.Getenv("INFLUXDB_ADDRESS"), influxdb.Database(os.Getenv("INFLUXDB_DATABASE")))),
		),
	}
}

func (s SumService) Sum(ctx context.Context, r *proxy.SumRequest) (*proxy.SumResponse, error) {
	return &proxy.SumResponse{
		Sum: r.A + r.B,
	}, nil
}

func (s SumService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	proxy.RegisterSumServiceServer(server.(*grpc.Server), serviceInstance.(proxy.SumServiceServer))

	return nil
}

func main() {
	// create instance of the service with proper configuration
	// addresses of external information are passed via environment variables
	s := NewSumService()
	// nice clenup
	defer s.Dispose()

	s.Log().InfoWithFields(log.LogFields{
		"name":    s.Info().Name,
		"version": s.Info().Version,
		"port":    s.Info().Port,
	}, "Service initialized")

	// for debugging purposes
	s.Log().SetLogLevel(log.DebugLogLevel)

	// get host address (with port)
	addr, _ := s.GetHostAddress()

	// register service in service discovery catalog
	s.Log().InfoWithFields(log.LogFields{
		"name":    s.Info().Name,
		"version": s.Info().Version,
		"address": addr,
	}, "Registering service in service discovery catalog")

	if err := s.Discovery().RegisterService(discovery.WithInfo(s.Info()), discovery.WithAddress(addr)); err != nil {
		s.Log().FatalWithFields(log.LogFields{"error": err}, "Cannot register service")
	}

	// start gRPC server
	server := gRPC.NewServer()
	server.StartRPCService(s)
}
