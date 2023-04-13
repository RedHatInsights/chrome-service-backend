# Cloud services config

## Why are these files here

Currently, HCC UI deployment is running in "hybrid mode". Some UIs are leveraging the new frontend operator and some are still using the legacy pipeline. Because in FEO all cloud services config files are handled by the chrome deployment pipeline, none of the files are deployed when a PR is merged inside the GH repository.

The modules and navigation files have been temporarily moved over to the Chrome service API to ensure normal updates. Once HCC is out of the "hybrid mode" (all UIs are using FEO), all navigation and module files will be defined in `frontend.yml` embedded within the UI codebase.

## What has changed

### No more environment-based branches

The env branches do not exist in this repository (and were annoying). Each environment has its own directory.

### Production environment(s) do not update automatically

Any changes made to `prod-stable` or `prod-beta` will not be deployed after merge. The chrome service image has to be bumped in `app-interface`. Prod env is not fully onboarded for CI/CD.

### The qaprodauth is not supported here

This API is not running on `qaprodauth` and can cause issues in these environments.

### No more main.yml changes

This is also a legacy file. We will work on adding support for `main.yml` or an alternative to keep the API docs component up to date.

### The `routes` attribute for modules has changed

The `routes` attribute defined in `fed-modules.json` can no longer be a simple array of strings.

```diff
- "routes": ["/foo/bar"]
+ "routes": [{"pathname": "/foo/bar"}]
```

## Updating files

Apart from the `routes` attribute, the format has stayed the same. The only difference is the file locations and the environments.

### Finding files

All CSC files are located under the `/static` directory. Based on the required environment, you can choose which files to change.

**Modules**

- `/static/stable/prod/modules/fed-modules.json` for prod-stable
- `/static/stable/stage/modules/fed-modules.json` for stage-stable
- `/static/beta/prod/modules/fed-modules.json` for prod-beta
- `/static/beta/stage/modules/fed-modules.json` for stage-beta

**Navigation files**

- `/static/stable/prod/navigation/*.-navigation.json` for prod-stable
- `/static/stable/stage/navigation/*.-navigation.json` for stage-stable
- `/static/beta/prod/navigation/*.-navigation.json` for prod-beta
- `/static/beta/stage/navigation/*.-navigation.json` for stage-beta

**Services**

- `/static/stable/prod/services/services.json` for prod-stable
- `/static/stable/stage/services/services.json` for stage-stable
- `/static/beta/prod/services/services.json` for prod-beta
- `/static/beta/stage/services/services.json` for stage-beta

### Serving files locally

There are two options to start an asset server for local development

### I have golang installed and running

If your machine has golang setup, you can run the following command:

```sh
dev-static
```
This command will start dev server on `http://localhost:8000`

You can adjust the server port by adding port argument:

```sh
make dev-static port=9999
```

This command will start dev server on `http://localhost:9999`

### I have Node.js installed and running

If your machine does not have golang or you don't want to use golang, there is an alternative using nodejs:

```sh
make dev-static-node
```

This command will start dev server on `http://localhost:8000`

You can adjust the server port by adding port argument:

```sh
make dev-static-node port=9999
```

This command will start dev server on `http://localhost:9999`

### Making changes

You can use the [CSC documentation](https://github.com/RedHatInsights/cloud-services-config#chromefed-modulesjson) for reference.
