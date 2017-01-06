package influxdb

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gkarlik/quark/metrics"
	"github.com/influxdata/influxdb/client/v2"
	"time"
)

type influxdbMetricsReporter struct {
	Client  client.Client
	Options Options
}

// Option represents function which is used to set metrics reporter options
type Option func(*Options)

// Options represents cofiguration options for metrics reporter
type Options struct {
	Address  string
	Username string
	Password string
	Database string
}

// Database allows to set database name
func Database(database string) Option {
	return func(o *Options) {
		o.Database = database
	}
}

// Username allows to set database user name
func Username(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// Password allows to set database password
func Password(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// NewMetricsReporter creates instance of metrics reported based on influxdb
func NewMetricsReporter(address string, opts ...Option) *influxdbMetricsReporter {
	options := new(Options)
	for _, o := range opts {
		o(options)
	}

	log.WithField("address", address).Info("Creating InfluxDB HTTP client")

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     address,
		Username: options.Username,
		Password: options.Password,
	})

	if err != nil {
		log.WithFields(log.Fields{
			"address":  address,
			"username": options.Username,
			"password": options.Password,
			"database": options.Database,
		}).Error("Cannot create InfluxDB HTTP client")
	}

	return &influxdbMetricsReporter{
		Options: *options,
		Client:  c,
	}
}

func (r influxdbMetricsReporter) Report(ms []metrics.Metric) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: r.Options.Database,
	})

	if err != nil {
		return err
	}

	for _, m := range ms {
		p, _ := client.NewPoint(
			m.Name,
			m.Tags,
			m.Values,
			time.Now(),
		)
		bp.AddPoint(p)
	}

	if err = r.Client.Write(bp); err != nil {
		return err
	}

	return nil
}

func (r influxdbMetricsReporter) Dispose() {
	if r.Client != nil {
		r.Client.Close()
	}
}