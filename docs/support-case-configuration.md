# Support case configuration

This configuration is added to a support case when the "createCase" Chrome UI method is called.

## Configuration

The configuration can be added to UI modules via:
1. global UI module configuration
2. UI module route configuration

If a module has bot global and route configuration, the route configuration will have a higher priority.

## Attributes

### Product

*string*
A human readable description of a product

### Version

*string*
Version of the support case

## Where to configure the support case data

The support case data is configured within `fed-mods.json` files.

### Global configuration example

```json
{
    "inventory": {
        "config": {
        "supportCaseData": {
                "product": "Red Hat Insights",
                "version": "Inventory"
            }
        },
    }
    // rest of the module config
}
```

### Route configuration example

```json
{
      "patch": {
        "config": {
            // will be ignored on /insights/patch/packages route because
            // this route has an explicit support case configuration as well
            "supportCaseData": {
                "product": "Red Hat Insights",
                "version": "Patch"
            }
        },
        "modules": [
            {
                "id": "patch",
                "module": "./RootApp",
                "routes": [
                    {
                        "pathname": "/insights/patch/packages",
                        // route config has higher priority
                        "supportCaseData": {
                            "product": "Red Hat Insights",
                            "version": "Content Services"
                        }
                    }
                ]
            }
        ]
    }
}
```

