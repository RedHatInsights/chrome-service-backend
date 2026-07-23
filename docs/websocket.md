# WebSockets in Chrome

WebSockets in chrome are implemented using Kafka as the underlying messaging service. You can use WebSockets 
in chrome to pull realtime data from the kafka stack. Messages are injected exclusively through the Kafka consumer path.

This document is a work in progress and will be updated as we add support for other environments.

## Getting Started

### Client Side Testing

WebSocket connection are provided by the gateway, so chrome does not handle the
connection upgrade directly. 

To open a new websocket locally, spin up chrome-service-backend open your browser to `http://localhost:8000`. Be sure to run `make infra` before testing.

Open the console, and add the following connection snippet.

`x = new WebSocket("wss://stage.foo.redhat.com:1337/wss/chrome-service/v1/ws", 'cloudevents.json')`

Once connected, you can log out any new wss messages with `x.onmessage = console.log`

To send test messages, use the Kafka consumer path. Run `go run cmd/kafka/testMessage.go` in a separate terminal (see "Reading Kafka Data" below).

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

## Reading Kafka Data

Ensure you have run `make infra` and that `docker ps` shows a running kafka.

Ensure that you have run `go run .` and that chrome-service has come up.

Run `go run cmd/kafka/testMessage.go` in a separate window. Once complete, you should see a new kafka message in chrome-service's logs. Feel free to adjust the script to change the test values as you wish. 

