package prometheus_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkarlik/quark-go/metrics/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewMetricsExposer(t *testing.T) {
	mex := prometheus.NewMetricsExposer(
		prometheus.Address("localhost:1234"),
		prometheus.EndPointName("/m"),
	)

	defer mex.Dispose()

	assert.Equal(t, "localhost:1234", mex.Options.Address)
	assert.Equal(t, "/m", mex.Options.EndPointName)
}

func TestMetricsExposer(t *testing.T) {
	mex := prometheus.NewMetricsExposer()
	defer mex.Dispose()

	ts := httptest.NewServer(
		mex.ExposeHandler(),
	)
	defer ts.Close()

	gauge := mex.CreateGauge("gauge", "gauge description")
	gauge.Set(1.2)

	counter := mex.CreateCounter("counter", "counter description")
	counter.Inc()

	histogram := mex.CreateHistogram("histogram", "histogram description", []float64{0.1, 0.5, 0.9})
	histogram.Observe(0.4)

	summary := mex.CreateSummary("summary", "summary description", map[float64]float64{0.1: 1.0})
	summary.Observe(0.5)

	response, err := http.Get(ts.URL)
	assert.NoError(t, err, "Error exposing metrics")

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err, "Error exposing metrics")

	assert.Contains(t, string(contents), "# HELP gauge gauge description")
	assert.Contains(t, string(contents), "# TYPE gauge gauge")
	assert.Contains(t, string(contents), "gauge 1.2")
	assert.Equal(t, "gauge", gauge.Name())
	assert.Equal(t, "gauge description", gauge.Description())

	assert.Contains(t, string(contents), "# HELP counter counter description")
	assert.Contains(t, string(contents), "# TYPE counter counter")
	assert.Contains(t, string(contents), "counter 1")

	assert.Contains(t, string(contents), "# HELP histogram histogram description")
	assert.Contains(t, string(contents), "TYPE histogram histogram")
	assert.Contains(t, string(contents), "histogram_bucket{le=\"0.1\"} 0")
	assert.Contains(t, string(contents), "histogram_bucket{le=\"0.5\"} 1")
	assert.Contains(t, string(contents), "histogram_bucket{le=\"0.9\"} 1")
	assert.Contains(t, string(contents), "histogram_bucket{le=\"+Inf\"} 1")
	assert.Contains(t, string(contents), "histogram_sum 0.4")
	assert.Contains(t, string(contents), "histogram_count 1")

	assert.Contains(t, string(contents), "# HELP summary summary description")
	assert.Contains(t, string(contents), "# TYPE summary summary")
	assert.Contains(t, string(contents), "summary{quantile=\"0.1\"} 0.5")
	assert.Contains(t, string(contents), "summary_sum 0.5")
	assert.Contains(t, string(contents), "summary_count 1")
}
