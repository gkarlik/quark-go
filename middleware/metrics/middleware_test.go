package metrics_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/metrics/prometheus"
	"github.com/gkarlik/quark-go/middleware/metrics"
	tr "github.com/gkarlik/quark-go/service/trace/noop"
	"github.com/stretchr/testify/assert"
)

type TestService struct {
	*quark.ServiceBase
}

type TestHttpHandler struct{}

func (h *TestHttpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestMetricsMiddleware(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Metrics(prometheus.NewMetricsExposer()),
			quark.Tracer(tr.NewTracer())),
	}
	defer ts.Dispose()

	mm := metrics.NewRequestMetricsMiddleware(ts)
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	h := mm.Handle(&TestHttpHandler{})
	h.ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMetricsMiddlewareWithNext(t *testing.T) {
	a, _ := quark.GetHostAddress(1234)

	ts := &TestService{
		ServiceBase: quark.NewService(
			quark.Name("TestService"),
			quark.Version("1.0"),
			quark.Address(a),
			quark.Metrics(prometheus.NewMetricsExposer()),
			quark.Tracer(tr.NewTracer())),
	}
	defer ts.Dispose()

	mm := metrics.NewRequestMetricsMiddleware(ts)
	r, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	th := &TestHttpHandler{}
	mm.HandleWithNext(w, r, th.ServeHTTP)

	assert.Equal(t, http.StatusOK, w.Code)
}
