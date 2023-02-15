# HTTP Server Example

This directory contains an example of how one might use this package in an HTTP
server. To run the example server:

```bash
go run github.com/Silicon-Ally/zapgcp/examples --local=$LOCAL --min_log_level=$LOG_LEVEL
```

Then you can hit various endpoints to see what logs are produced, ex.

```bash
curl localhost:8080/
curl localhost:8080/some/path
curl localhost:8080/admin/test
```

Try changing `$LOCAL` from `true` to `false` to see how the logs change, or set
`$LOG_LEVEL` to `warn` or `error` to only see higher log levels. `8080` is the
default port and can be changed with the `--port` flag.
