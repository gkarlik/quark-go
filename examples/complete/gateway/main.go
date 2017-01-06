package main

import (
	"fmt"
	"github.com/gkarlik/quark"
	"github.com/gkarlik/quark/auth/jwt"
	"github.com/gkarlik/quark/broker"
	"github.com/gkarlik/quark/broker/rabbitmq"
	proxy "github.com/gkarlik/quark/examples/complete/service/definitions/proxies/sum"
	"github.com/gkarlik/quark/logger"
	"github.com/gkarlik/quark/logger/logrus"
	"github.com/gkarlik/quark/metrics/influxdb"
	"github.com/gkarlik/quark/ratelimiter"
	"github.com/gkarlik/quark/service/discovery"
	"github.com/gkarlik/quark/service/discovery/consul"
	"github.com/gkarlik/quark/service/loadbalancer/random"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Gateway struct {
	*quark.ServiceBase
}

func NewGateway() *Gateway {
	port, _ := strconv.Atoi(os.Getenv("SERVICE_PORT"))

	return &Gateway{
		ServiceBase: quark.NewService(
			quark.Name(os.Getenv("SERVICE_NAME")),
			quark.Version(os.Getenv("SERVICE_VERSION")),
			quark.Tags("A"),
			quark.Port(port),
			quark.Logger(logrus.NewLogger()),
			quark.Discovery(consul.NewServiceDiscovery(os.Getenv("CONSUL_ADDRESS"))),
			quark.Broker(rabbitmq.NewMessageBroker(os.Getenv("RABBITMQ_ADDRESS"))),
			quark.Metrics(influxdb.NewMetricsReporter(os.Getenv("INFLUXDB_ADDRESS"), influxdb.Database(os.Getenv("INFLUXDB_DATABASE")))),
		),
	}
}

var gateway *Gateway

func main() {
	gateway = NewGateway()
	gateway.Log().Info("Gateway initialized")

	r := mux.NewRouter()

	r.HandleFunc("/sum/{a}/{b}", sumHandler)
	r.HandleFunc("/message/{key}/{msg}", messageHandler)

	addr, _ := gateway.GetHostAddress()
	gateway.Log().InfoWithFields(logger.LogFields{
		"address": addr,
	}, "Listening on address")

	limiter := ratelimiter.NewHTTPRateLimiter(10 * time.Second)
	auth := jwt.NewAuthenticationMiddleware()

	http.ListenAndServe(fmt.Sprintf(":%d", gateway.Info().Port), auth.Handle(limiter.Handle(r)))
}

func findSumService() *grpc.ClientConn {
	gateway.Log().Info("Locating SumService")

	addr, err := gateway.Discovery().GetServiceAddress(
		discovery.ByName("SumService"),
		discovery.ByTag("A"),
		discovery.UsingLBStrategy(random.NewRandomLBStrategy()),
	)

	if err != nil {
		gateway.Log().Error(err)
	}

	conn, err := grpc.Dial(addr.String(), grpc.WithInsecure())
	if err != nil {
		gateway.Log().FatalWithFields(logger.LogFields{
			"error": err,
		}, "Cannot connect to server")
	}
	return conn
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	msg := vars["msg"]

	err := gateway.Broker().PublishMessage(broker.Message{
		Key:   key,
		Value: msg,
	})

	if err != nil {
		w.WriteHeader(501)
		fmt.Fprintf(w, "Error: %v", err)

		return
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "Message successfully sent")
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	a, _ := strconv.ParseInt(vars["a"], 10, 64)
	b, _ := strconv.ParseInt(vars["b"], 10, 64)

	conn := findSumService()
	defer conn.Close()

	client := proxy.NewSumServiceClient(conn)
	result, err := client.Sum(context.Background(), &proxy.SumRequest{A: a, B: b})

	if err != nil {
		w.WriteHeader(501)
		fmt.Fprintf(w, "Error: %v", err)

		return
	}
	w.WriteHeader(200)
	fmt.Fprintf(w, "%d + %d = %d", a, b, result.Sum)
}
