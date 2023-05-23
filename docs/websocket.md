# WebSockets in Chrome

WebSockets in chrome are implemented using Kafka as the underlying messaging service. You can use WebSockets 
in chrome to pull realtime data from the kafka stack, as well as send messages. 

This document is a work in progress and will be updated as we add support for other environments.

## Getting Started

### Client Side Testing

WebSocket connection are provided by the gateway, so chrome does not handle the
connection upgrade directly. 

To open a new websocket locally, spin up chrome-service-backend open your browser to `http://localhost:8000`. Be sure to run `make infra` before testing.

Open the console, and add the following connection snippet.

`x = new WebSocket("ws://localhost:8000/wss/chrome-service/v1/ws");`

Once connected, you can log out any new wss messages with `x.onmessage = console.log`

You can send a test message with the following fetch request. Please note the 
header is only a test value and will not work in any other environment.

**When testing in chrome environment, make sure to allow the endpoint in the CSP headers in index.html!**

``` javascript
fetch("/api/chrome-service/v1/emit-message", {
  method: "POST",
  body: JSON.stringify({
    payload: { foo: "bar" },
  }),
  headers: {
    "x-rh-identity": "eyJpZGVudGl0eSI6eyJ1c2VyIjp7InVzZXJfaWQiOiIxMiJ9fX0="
  }
});
```

You should see data come into the console logs on the chrome-service terminal window.

## Reading Kafka Data

Ensure you have run `make infra` and that `docker ps` shows a running kafka.

Ensure that you have run `go run .` and that chrome-service has come up.

Run `go run cmd/kafka/testMessage.go` in a separate window. Once complete, you should see a new kafka message in chrome-service's logs. Feel free to adjust the script to change the test values as you wish. 

