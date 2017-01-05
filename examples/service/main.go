package main

import (
	"carlos/quark"
	"carlos/quark/service"
	"carlos/quark/service/bus"
	"carlos/quark/service/bus/rabbitmq"
	"carlos/quark/service/discovery"
	"carlos/quark/service/discovery/consul"
	"carlos/quark/service/lb/random"
	"carlos/quark/service/log"
	"carlos/quark/service/log/logrus"
	"carlos/quark/service/metrics"
	"carlos/quark/service/metrics/influxdb"
	"carlos/quark/service/rpc/grpc"
	"golang.org/x/net/context"
	"time"
)

type SimpleService struct {
	*quark.ServiceBase
}

func NewSimpleService(consulAddress string, rabbitMQAddress string, influxdbAddress string) *SimpleService {
	return &SimpleService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Port(9999),
			quark.Logger(logrus.NewLogrusLogger()),
			quark.Discovery(consul.NewConsulServiceDiscovery(consulAddress)),
			quark.Bus(rabbitmq.NewRabbitMQBus(rabbitMQAddress)),
			quark.Metrics(influxdb.NewInfluxdbMetricsReporter(influxdbAddress, influxdb.Database("test"))),
		),
	}
}

func (s SimpleService) Sum(ctx context.Context, a int, b int) int {
	return a + b
}

func (s SimpleService) RegisterServiceInstance(server interface{}, serviceInstance interface{}) error {
	return nil
}

func main() {
	s := NewSimpleService("consul:8080", "amqp://rabitmq:9090", "http://influxdb:1111")
	defer s.Dispose()

	s.Log().SetLogLevel(log.DebugLogLevel)

	s.Log().Debug("test")

	sum := s.Sum(context.Background(), 1, 2)
	s.Log().DebugWithFields(log.LogFields{"sum": sum}, "Sum is")

	err := s.Discovery().RegisterService(
		discovery.WithInfo(s.Info()),
		discovery.WithAddress(service.NewURIServiceAddress("http://localhost:8080")))

	if err != nil {
		s.Log().Error(err)
	}

	a, err := s.Discovery().GetServiceAddress(
		discovery.ByName("Test"),
		discovery.ByTag("A"),
		discovery.ByInfo(s.Info()),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()),
	)

	if err != nil {
		s.Log().Panic(err)
	}

	s.Metrics().Report([]metrics.Metric{metrics.Metric{
		Date: time.Now(),
		Name: "test",
		Values: map[string]interface{}{
			"a": 1,
		},
		Tags: map[string]string{
			"a": "1",
			"b": "2",
		},
	}})

	s.Log().DebugWithFields(log.LogFields{"address": a.String(), "name": "Test"}, "Address of service")

	s.Bus().PublishMessage(bus.Message{
		Key:   "test",
		Value: "dupa",
	})

	msgs, _ := s.Bus().Subscribe("test")
	m := <-msgs

	s.Log().DebugWithFields(log.LogFields{"k": m.Key, "v": m.Value}, "Message arrived")

	server := grpc.NewGRPCServer()
	server.StartRPCService(s)
}
