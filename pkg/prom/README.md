# go-monitoring

This library can be used to instrument Golang services for Prometheus scraping
(http server- and client-side instrumentation)

## Installation

Import `github.com/gesundheitscloud/go-svc/pkg/prom` and run `go mod tidy`.

## Server-side instrumentation

If the service exposes http endpoints, this library helps instrumenting all incoming http requests with the following metrics under the given name (<subsystem> defaults to `phdp`). Namespace defaults to `d4l` and can be changed via `prom.SetNamespace("myapp")`:

- [HTTP Request Gauge Metric](https://prometheus.io/docs/concepts/metric_types/#gauge): `<namespace>_<subsystem>_http_requests`
- [HTTP Request Counter Metric](https://prometheus.io/docs/concepts/metric_types/#counter): `<namespace>_<subsystem>_http_requests_total`
- [HTTP Request Duration Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `<namespace>_<subsystem>_http_request_duration_seconds`
- Optional: [HTTP Request Size Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `<namespace>_<subsystem>_http_request_size_bytes`
- Optional: [HTTP Response Size Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `<namespace>_<subsystem>_http_response_size_bytes`

It uses the following default buckets for histograms (changeable through the `prom.NewHandlerInstrumenter(...)` init options):

- Size buckets: `[]float64{1024, 5120, 20480, 102400, 512000, 1048576, 10485760, 52428800} // 1KB, 5KB, 20KB, 100KB, 500KB, 1MB, 10MB, 50MB`
- Latency buckets: `[]float64{.25, .5, 1, 2.5, 5, 10} // seconds`

### Usage (server-side)

This library provides an [http handler](https://golang.org/pkg/net/http/#Handler) chain and can be used as follows:

``` golang
package handler

import (
   "net/http"

   prom "github.com/gesundheitscloud/go-svc/pkg/prom"

   "github.com/gorilla/mux"
   "github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler() http.Handler {
   router := mux.NewRouter()
   prom.SetNamespace("myapp") // optional; defaults to d4l
   mon := prom.NewHandlerInstrumenter(prom.WithSubsystem("myOwnSubSystem")) // defaults to phdp subsystem

   router.
      Methods(http.MethodGet).
      Path("/users/{owner:[A-Fa-f0-9-]+}/records/{record:[A-Fa-f0-9-]+}").
      Handler(mon.Instrument("/users/id/records/id",
         myHandler,
         prom.WithRespSize))

   router.Methods(http.MethodGet).Path("/metrics").Handler(promhttp.Handler()) // important to expose /metrics to enable prometheus scraping

   return router
}
```

## Client-side instrumentation

If the service executes outgoing http requests, this library helps instrumenting all outgoing http requests with the following metrics under the given name (<subsystem> defaults to `phdp`). Namespace can be changed via `prom.SetNamespace(...)`:

- [HTTP Request Gauge Metric](https://prometheus.io/docs/concepts/metric_types/#gauge): `<namespace>_<subsystem>_http_out_requests`
- [HTTP Request Counter Metric](https://prometheus.io/docs/concepts/metric_types/#counter): `<namespace>_<subsystem>_http_out_requests_total`
- [HTTP Request Duration Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `<namespace>_<subsystem>_http_out_request_duration_seconds`

It uses the same size bucket as described above for the server-side instrumentation.

### Usage (client-side)

This library provides an [http roundtripper](https://golang.org/pkg/net/http/#RoundTripper) and can be used as follows (this example also includes the logging roundtripper provided by `go-log`):

``` golang
func newClient() *http.Client {
   return &http.Client{
       Transport: log.LoggedTransport(
           mon.Instrument("myClientName",
              http.DefaultTransport)),
       // http default client has zero timeout (no timeout).
       // Set to 30 sec to mitigate potential network issues
       Timeout: 30 * time.Second,
    }
}
```
