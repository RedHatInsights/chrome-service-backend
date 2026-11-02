# WebSockets in Chrome

WebSockets in chrome are implemented using Kafka as the underlying messaging service. You can use WebSockets 
in chrome to pull realtime data from the kafka stack. Messages are published to a Kafka topic by upstream services, and the server-side Kafka consumer reads them and forwards them to connected WebSocket clients via the connection hub.

This document is a work in progress and will be updated as we add support for other environments.

## Getting Started

### Client Side Testing

WebSocket connection are provided by the gateway, so chrome does not handle the
connection upgrade directly. 

To open a new websocket locally, spin up chrome-service-backend open your browser to `http://localhost:8000`. Be sure to run `make infra` before testing.

Open the console, and add the following connection snippet.

`x = new WebSocket("wss://stage.foo.redhat.com:1337/wss/chrome-service/v1/ws", 'cloudevents.json')`

Once connected, you can log out any new wss messages with `x.onmessage = console.log`

To send test messages, run `go run cmd/kafka/testMessage.go` in a separate terminal — this publishes a test message to the Kafka topic, which the server-side consumer then delivers to the WebSocket (see "Reading Kafka Data" below).

**Make sure your chrome has ws proxy setup for the endpoint! This is required to have proper WS registration.**

Sample proxy config:
```jsx
{ 
  routes:
    {
    ...(process.env.CHROME_SERVICE && {
      // web sockets
      '/wss/chrome-service/': {
        target: `ws://localhost:${process.env.CHROME_SERVICE}`,
        // To upgrade the connection
        ws: true,
      },
      // REST API
      '/api/chrome-service/v1/': {
        host: `http://localhost:${process.env.CHROME_SERVICE}`,
      },
    })
  }
}
```

You should see data come into the console logs on the chrome-service terminal window.

## Diagnosing rejected origins

An origin rejection emits one JSON `websocket_origin_rejected` event (rather than a second generic upgrade-error event). It includes `request_id`, `origin_header_state` (`missing`, `null`, `empty`, or `present`), a safe `origin_representation`, `origin_rejection_reason` (`missing`, `malformed`, `disallowed_scheme`, or `disallowed_host`), request method/host/path, bounded User-Agent, HTTP status, and `upgrade_error_category`. Origin userinfo, path, query, and fragment are never logged. Valid DNS hostnames may appear; IP literals, unsafe hostnames, and hosts matching credential markers are redacted. Credentials, cookies, identity headers, and IP addresses are excluded. Other WebSocket handshake failures remain logged as `upgrade_failed`.

Example CloudWatch Logs Insights query:

```text
fields @timestamp, request_id, origin_rejection_reason, origin_header_state,
       origin_representation, method, host, path, status, upgrade_error_category
| filter event = "websocket_origin_rejected"
| sort @timestamp desc
```

Use `request_id` to correlate the rejection with adjacent request or gateway logs. To identify the running environment and build, use the log group/stream and deployment-platform metadata already attached to the selected log record or workload: environment/namespace, pod, image/tag or digest, and commit label (Konflux build metadata uses `build.appstudio.redhat.com/commit_sha`). Metadata availability varies by environment; do not infer a deployment or commit when those fields are absent. This event alone does not prove which image is deployed.

## Reading Kafka Data

Ensure you have run `make infra` and that `docker ps` shows a running kafka.

Ensure that you have run `go run .` and that chrome-service has come up.

Run `go run cmd/kafka/testMessage.go` in a separate window. Once complete, you should see a new kafka message in chrome-service's logs. Feel free to adjust the script to change the test values as you wish. 
