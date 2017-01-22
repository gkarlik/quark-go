package influxdb

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/gkarlik/quark/metrics"
	"github.com/influxdata/influxdb/client/v2"
)

// MetricsReporter represents kpi reporting mechanism based on InfluxDB
type MetricsReporter struct {
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

// NewMetricsReporter creates instance of metrics reported based on influxdb. Panics if cannot create an instance
func NewMetricsReporter(address string, opts ...Option) *MetricsReporter {
	options := new(Options)
	for _, o := range opts {
		o(options)
	}

	options.Address = address

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
			"error":    err,
		}).Panic("Cannot create InfluxDB HTTP client")
	}

	return &MetricsReporter{
		Options: *options,
		Client:  c,
	}
}

// Report send metrics to InfluxDB
func (r MetricsReporter) Report(ms []metrics.Metric) error {
	if ms == nil || len(ms) == 0 {
		return errors.New("Metrics array cannot be nil or empty")
	}

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
			m.Date,
		)
		bp.AddPoint(p)
	}

	if err = r.Client.Write(bp); err != nil {
		return err
	}

	return nil
}

// Dispose cleans up MetricsReporter instance
func (r MetricsReporter) Dispose() {
	if r.Client != nil {
		r.Client.Close()
	}
}
