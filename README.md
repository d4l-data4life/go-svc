# Boilerplate code for golang services @ D4L

This repository serves as the the "library monorepo" for golang services at data4life
and comprises the following packages:

- [bievents](./pkg/bievents): Instrumentation of events relevant for BI reporting
- [client](./pkg/client)
- [cors](./pkg/cors): Handling of Cross-Origin Resource Sharing headers
- [d4lcontext](./pkg/d4lcontext)
- [db](./pkg/db)
- [instrumented](./pkg/instrumented)
- [jwt](./pkg/jwt): Handling of JWTs
- [log](./pkg/log): Logging and Audit Logging
- [logging](./pkg/logging)
- [middlewares](./pkg/middlewares)
- [migrate](./pkg/migrate): Schema creation and migration (DDL) for PostgreSQL databases
- [probe](./pkg/probe)
- [prom](./pkg/prom): Prometheus monitoring instrumentation
- [standard](./pkg/standard)
