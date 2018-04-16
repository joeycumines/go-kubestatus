# kubestatus.go

Simplistic opinionated HTTP status server library designed to be provided 
a bare minimum of configuration (intended for Kubernetes).

https://godoc.org/github.com/joeycumines/kubestatus.go

- Sets up a HTTP server that serves `/readiness` and `/healthz` endpoints
- Provides a client implementation supporting multiple instances
- Configurable bind port / hostname (uses https://github.com/gin-gonic/gin)
- Allows overriding gin handlers
- The status of both endpoints are defined by callbacks in the form `func() error`
- Standard response object documented by [swagger.yml](swagger.yml), which includes UUID
    for the process
- The `/readiness` endpoint can be automatically extended to check for the `/readiness` of 
    other services via simple dependency config
- Circular reference detection is in-built for `/readiness`, and automatically wired up
    if used via the dependency config
- Exposes a context and fatal error info that can be used to handle fatal errors with the 
    server including panics

The tests are very bad but complete-ish.
