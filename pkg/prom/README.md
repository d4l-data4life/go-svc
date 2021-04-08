# go-monitoring

This library can be used to instrument Golang services for Prometheus scraping
(http server- and client-side instrumentation)

## Installation

The package should be installed **automatically** after running any `go` command, provided that the code imports and uses this package.

When installing a package from private repositories, Go needs to be instructed to not validate the checksums in a public database.
This can be obtained by setting the following environmental variable.

```bash
export GOPRIVATE='github.com/gesundheitscloud/*'
# or
go env -w GOPRIVATE='github.com/gesundheitscloud/*'
```

Then installation can be restarted with:

```bash
go mod tidy
# or explicitly
go get github.com/gesundheitscloud/go-monitoring@v0.2.0
```

Should the environment be set incorrectly, the following error may occur:

```
verifying github.com/gesundheitscloud/go-monitoring@v0.2.0: github.com/gesundheitscloud/go-monitoring@v0.2.0: reading https://sum.golang.org/lookup/github.com/gesundheitscloud/go-monitoring@v0.2.0: 410 Gone
```

## Server-side instrumentation

If the service exposes http endpoints, this library helps instrumenting all incoming http requests with the following metrics under the given name (\<subsystem\> defaults to `phdp`):

- [HTTP Request Gauge Metric](https://prometheus.io/docs/concepts/metric_types/#gauge): `d4l_<subsystem>_http_requests`
- [HTTP Request Counter Metric](https://prometheus.io/docs/concepts/metric_types/#counter): `d4l_<subsystem>_http_requests_total`
- [HTTP Request Duration Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `d4l_<subsystem>_http_request_duration_seconds`
- Optional: [HTTP Request Size Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `d4l_<subsystem>_http_request_size_bytes`
- Optional: [HTTP Response Size Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `d4l_<subsystem>_http_response_size_bytes`

It uses the following default buckets for histograms (changeable through the `prom.NewHandlerInstrumenter(...)` init options):

- Size buckets: `[]float64{1024, 5120, 20480, 102400, 512000, 1048576, 10485760, 52428800} // 1KB, 5KB, 20KB, 100KB, 500KB, 1MB, 10MB, 50MB`
- Latency buckets: `[]float64{.25, .5, 1, 2.5, 5, 10} // seconds`

### Usage (server-side)

This library provides an [http handler](https://golang.org/pkg/net/http/#Handler) chain and can be used as follows:

``` golang
package handler

import (
   "net/http"

   "github.com/gesundheitscloud/go-monitoring/prom"

   "github.com/gorilla/mux"
   "github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler() http.Handler {
   router := mux.NewRouter()
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

If the service executes outgoing http requests, this library helps instrumenting all outgoing http requests with the following metrics under the given name (\<subsystem\> defaults to `phdp`):

- [HTTP Request Gauge Metric](https://prometheus.io/docs/concepts/metric_types/#gauge): `d4l_<subsystem>_http_out_requests`
- [HTTP Request Counter Metric](https://prometheus.io/docs/concepts/metric_types/#counter): `d4l_<subsystem>_http_out_requests_total`
- [HTTP Request Duration Histogram](https://prometheus.io/docs/concepts/metric_types/#histogram): `d4l_<subsystem>_http_out_request_duration_seconds`

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
