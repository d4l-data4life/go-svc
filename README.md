[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/d4l-data4life/go-svc)



## Go service toolkit

Reusable building blocks for Go services: logging, middleware, database helpers,
Prometheus instrumentation, HTTP client/server utilities, and more.

### Packages

- `pkg/client`: HTTP client helpers and OAuth2 client
- `pkg/db`: GORM setup, connection management, and metrics
- `pkg/instrumented`: Handler factory with structured logging and metrics
- `pkg/log`: Structured logging, audit logs, HTTP request/response logging
- `pkg/logging`: Global logger facade for convenience
- `pkg/middlewares`: Auth, tenant, tracing, URL filter middlewares
- `pkg/migrate`: Migration runner for PostgreSQL (SQL files)
- `pkg/prom`: Prometheus metrics utilities for HTTP client/server
- `pkg/standard`: Opinionated server/gateway wiring
- `pkg/ticket`: Lightweight JWT ticket verification/claims
- `pkg/transport`: Composable RoundTripper chain (retry, timeout, auth, trace)

### Prometheus namespace

Metrics default to the `d4l` namespace for backward compatibility. You can set a
custom namespace at runtime:

```go
import (
    dbMetrics "github.com/d4l-data4life/go-svc/pkg/db"
    prom "github.com/d4l-data4life/go-svc/pkg/prom"
)

func init() {
    prom.SetNamespace("myapp")
    dbMetrics.SetPrometheusNamespace("myapp")
}
```

### License

Apache License 2.0. See `LICENSE`.
