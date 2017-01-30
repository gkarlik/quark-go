package influxdb_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gkarlik/quark-go/metrics"
	"github.com/gkarlik/quark-go/metrics/influxdb"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsReporter(t *testing.T) {
	mr := influxdb.NewMetricsReporter("http://influxdb/",
		influxdb.Database("database"),
		influxdb.Username("user"),
		influxdb.Password("password"))

	defer mr.Dispose()

	assert.Equal(t, "http://influxdb/", mr.Options.Address)
	assert.Equal(t, "database", mr.Options.Database)
	assert.Equal(t, "user", mr.Options.Username)
	assert.Equal(t, "password", mr.Options.Password)
}

func TestMetricsReporter(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))
	defer ts.Close()

	mr := influxdb.NewMetricsReporter(ts.URL,
		influxdb.Database("database"),
		influxdb.Username("user"),
		influxdb.Password("password"))

	m1 := metrics.Metric{
		Date: time.Date(2017, time.January, 1, 18, 4, 20, 0, time.UTC),
		Name: "test",
		Tags: map[string]string{
			"A": "1",
			"B": "2",
		},
		Values: map[string]interface{}{
			"C": 3,
			"D": 4,
		},
	}

	m2 := metrics.Metric{
		Date: time.Date(2017, time.January, 1, 10, 10, 0, 0, time.UTC),
		Name: "test1",
		Tags: map[string]string{
			"A": "5",
			"B": "6",
		},
		Values: map[string]interface{}{
			"C": 7,
			"D": 8,
		},
	}

	err := mr.Report(m1, m2)

	assert.NoError(t, err, "Error reporting metrics")
	assert.Equal(t, "/write?consistency=&db=database&precision=ns&rp=", data.url)
	assert.Equal(t, "test,A=1,B=2 C=3i,D=4i 1483293860000000000\ntest1,A=5,B=6 C=7i,D=8i 1483265400000000000\n", data.body)
}

func TestMetricsReporterNetworkError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	mr := influxdb.NewMetricsReporter(ts.URL,
		influxdb.Database("database"),
		influxdb.Username("user"),
		influxdb.Password("password"))

	m1 := metrics.Metric{
		Date: time.Date(2017, time.January, 1, 18, 4, 20, 0, time.UTC),
		Name: "test",
		Tags: map[string]string{
			"A": "1",
			"B": "2",
		},
		Values: map[string]interface{}{
			"C": 3,
			"D": 4,
		},
	}

	err := mr.Report(m1)

	assert.Error(t, err, "Report should return an error")
}

func TestMetricsReporterEmptyDate(t *testing.T) {
	data := struct {
		url  string
		body string
	}{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)

		b, _ := ioutil.ReadAll(r.Body)

		data.url = r.URL.String()
		data.body = string(b)
	}))
	defer ts.Close()

	mr := influxdb.NewMetricsReporter(ts.URL,
		influxdb.Database("database"),
		influxdb.Username("user"),
		influxdb.Password("password"))

	m1 := metrics.Metric{
		Name: "test",
		Tags: map[string]string{
			"A": "1",
			"B": "2",
		},
		Values: map[string]interface{}{
			"C": 3,
			"D": 4,
		},
	}

	err := mr.Report(m1)

	assert.NoError(t, err, "Error reporting metrics")
	assert.Equal(t, "/write?consistency=&db=database&precision=ns&rp=", data.url)
	assert.Contains(t, data.body, "test,A=1,B=2 C=3i,D=4i")
}

func TestMetricsEmptyList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	mr := influxdb.NewMetricsReporter(ts.URL,
		influxdb.Database("database"),
		influxdb.Username("user"),
		influxdb.Password("password"))

	err := mr.Report()

	assert.Error(t, err, "Report should return an error")
}

func TestMetricsReporterClientError(t *testing.T) {
	assert.Panics(t, func() {
		influxdb.NewMetricsReporter("incorrect url",
			influxdb.Database("database"),
			influxdb.Username("user"),
			influxdb.Password("password"))
	})
}
