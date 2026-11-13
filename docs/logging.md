# Structured logging

The service emits JSON logs through logrus. CloudWatch is used when configured;
otherwise logs go to stdout. CloudWatch hooks are flushed during normal shutdown
and before a fatal startup/runtime exit.

## Levels and HTTP completion events

`LOG_LEVEL` sets the minimum emitted level. Supported values are logrus levels
(`panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`); an unset or invalid
value falls back to `error`.

HTTP completion events use status-based severity:

| Status | Event level | Included when `LOG_LEVEL` is |
|---|---|---|
| 5xx | `error` | `error` or more verbose |
| 4xx | `warning` | `warn` or more verbose |
| Other | `info` | `info` or more verbose |

Completion fields: `request_id`, `method`, `route` (chi route pattern, or
`unmatched`), `status`, `bytes`, `elapsed_ms`, and available parsed identity
fields (`user_id`, `org_id`, `account`). Request paths, query strings, remote
addresses, headers, cookies, request/response bodies, and identity tokens are
not included in completion events.

Security events retain their schema: `event`, `action`, `resource_type`,
`resource_id`, `outcome`, and available `user_id`/`org_id` principal fields.
Request-scoped security events also include `request_id`. Successful graceful
shutdown is an `info`/`success` event; an unexpected server or shutdown failure
is an `error`/`failure` event.

Recovered request panics emit `request panic recovered` with `request_id`,
available identity fields, `panic_type`, and a stack trace bounded to 8 KiB.
Panic values are not logged. Database diagnostics include a bounded operation
type, row count, elapsed time, and error type; SQL text and bound parameters are
never logged. Slow database operations are warned after 200 ms.

Kafka diagnostics use configured `topic`, `offset`, message `byte_len`, and the
`broadcast` flag where applicable. Message keys and payloads are never logged.

## CloudWatch Logs Insights examples

Find server errors with request correlation:

```text
fields @timestamp, level, message, request_id, method, route, status, elapsed_ms, user_id, org_id
| filter message = "request completed" and status >= 500
| sort @timestamp desc
| limit 100
```

Find client errors:

```text
fields @timestamp, request_id, method, route, status, elapsed_ms
| filter message = "request completed" and status >= 400 and status < 500
| sort @timestamp desc
| limit 100
```

Search a request's related request, security, and panic events by `request_id`.
Process startup/shutdown events have no request ID.
