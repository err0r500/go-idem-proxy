# Go Idem Proxy

an HTTP proxy to get idempotency on POST requests

Run
```
go-idem-proxy --redis-url=redis:6379 --target-url=http://localhost:3000
```

- The proxy will expect an `X-idem-token` header (can be overriden with `IDEM_TOKEN=myHeader` env var) on every POST request

if it's not present :
- it will reject the request with a 400
- Otherwise, it will check if there's something at the key=<value_of_the_header> (in the Redis database)

If there's something:
- it will return the cached response without forwarding the request
- Otherwise, it will forward the request and store the response for later use.

NB : the default `TTL` is set at 60 seconds, you can override that with `CACHE_TTL` env var

## Run Unit tests
```
go test ./...
```

## Run Integration & Unit tests
```
go test ./... -tags=integration
```
