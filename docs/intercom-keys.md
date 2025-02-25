# Adding or Updating Intercom Keys in chrome-service

## Overview

Some teams wish to use Intercom Messenger for their apps and will need API secrets added. This requires adding the keys themselves into the vault and then exposing environment variables that point to them in `chrome-service`.

## Adding secrets to the vault

The secrets themselves will go in 

https://vault.devshift.net/ui/vault/secrets/insights/show/secrets/insights-dev/platform-experience-dev/chrome-service-backend 

typically with one for stage (denoted by the `_DEV` ending) and one for prod (without `_DEV`). These secret names will be added to `config.go` and `clowdapp.yml` below.

## Updating chrome-service

In `config.go`, you need to update `options.IntercomConfig` with the environment variable names and their corresponding secret names in the vault. **The attribute name must match the UI module name!** This can be checked from the fed-modules.json config. In `clowdapp.yml`, you need to update `env` with the secret names that point to our `chrome-service-backend` in the vault. Also in the file, dummy secrets need to be added to `data` for CI/CD purposes (you can just copy the ones that are already provided).

## Bumping the secrets version in app-interface

Update [this line](https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/chrome-service/namespaces/chrome-service-stage.yml#L38) to update secrets in stage with the new secrets version (you may need to message someone if you don't have vault permissions to see file versioning).
