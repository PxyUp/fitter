package http_client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func observations(t *testing.T, host, method, status string) uint64 {
	t.Helper()
	h, ok := requestSeconds.WithLabelValues(host, method, status).(prometheus.Histogram)
	if !ok {
		t.Fatal("series is not a histogram")
	}
	m := &dto.Metric{}
	if err := h.Write(m); err != nil {
		t.Fatal(err)
	}
	return m.GetHistogram().GetSampleCount()
}

// A request through the default client lands in the histogram under its
// host, method and status — the three things a slow or failing scrape is
// diagnosed by.
func TestRequestsAreObservedByHostMethodAndStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	t.Cleanup(ts.Close)
	host := ts.Listener.Addr().String()

	okBefore := observations(t, host, "GET", "200")
	teapotBefore := observations(t, host, "POST", "418")

	client := GetDefaultClient()
	res, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	res, err = client.Post(ts.URL, "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	if got := observations(t, host, "GET", "200"); got != okBefore+1 {
		t.Errorf("GET 200 observations = %d, want %d", got, okBefore+1)
	}
	if got := observations(t, host, "POST", "418"); got != teapotBefore+1 {
		t.Errorf("POST 418 observations = %d, want %d", got, teapotBefore+1)
	}
}

// A connection that never reaches a server is still a data point — the
// scrape that hangs or refuses is exactly the one worth seeing.
func TestTransportErrorsAreObservedAsError(t *testing.T) {
	host := "127.0.0.1:1"
	before := observations(t, host, "GET", "error")

	_, err := GetDefaultClient().Get("http://" + host)
	if err == nil {
		t.Fatal("a connection to a closed port succeeded")
	}
	if got := observations(t, host, "GET", "error"); got != before+1 {
		t.Errorf("error observations = %d, want %d", got, before+1)
	}
}

// A request with no URL is answered by the underlying transport with an
// error, not a panic — and observing it must keep that true. The wrapper is
// never allowed to be less safe than what it wraps.
func TestANilURLIsAnErrorNotAPanic(t *testing.T) {
	req := &http.Request{Method: "GET", URL: nil}
	_, err := WrapTransport(nil).RoundTrip(req)
	if err == nil {
		t.Fatal("a request with no URL succeeded")
	}
}
