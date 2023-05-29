# Using Feature Flags in chrome-service

## Overview

Feature flags in chrome-service are based on the [unleash library](http://getunleash.io). We use the 
[unleash-client-go](https://github.com/Unleash/unleash-client-go) to access the server in our go client code.

## Setup

### Local

In order to use feature flags through unleash, we need to run a server. Since we're on local dev here, that's our problem.
We can make our own local unleash instance with `make infra` to spin up all the needed containers in docker-compose. If, 
for some reason, you want to only spin up unleash, `make unleash` will give you the containers you seek.

Navigate to http://localhost:4242 and login to unleash with username `admin` and password `unleash4all`. 

Create a new feature flag as you would like (we try to follow a pattern of `serviceName.feature.enabled` or something similar).

As you run `go run main.go` or `make dev`, unleash will be auto configured as long as you've set your .env file correctly.

### Ephemeral

Ephemeral is very similar, but your ClowdApp must include `featureFlags: true` in the yaml. Without this, the unleash server
will not be created. When your ephemeral environment is ready, you can find the created route in openshift or using `oc get route`.
Navigate to that route and login with the same flow from "Local" using `admin:unleash4all`. 

The pod should start with the unleash client already configured since the environment and secrets are fetched from Clowder.

### Stage / Prod

In order to use feature flags in stage or prod, you must gain access to the insights unleash instance through app-interface.

Once you get access and your Github Auth is enabled, you can add and toggle feature flags as detailed above. 

Again because the client gets its configuration from Clowder, you will not need to do any additional config work.

### Usage

In your Go code, the unleash client should already be setup so you can use `featureflags.IsEnabled("serviceName.feature.enabled)`
to read from the server. You will have control of the toggle as you please through the unleash UI.

