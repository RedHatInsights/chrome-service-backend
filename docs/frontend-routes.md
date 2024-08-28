# Frontend routes

Frontend routes are defined in UI module configuration files.

## Routes visibility

Similar to navigation items or search index entries, routes can leverage the `permissions` configuration to further restrict routes visibility on additional conditions rather than just environment. If a route permission is not met, a not found screen id displayed to the user.

### Example configuration

The following route will be visible only for users which have @foo.bar email. Any chrome [visibility method](https://github.com/RedHatInsights/insights-chrome/blob/master/docs/navigation.md#permissions) can be used.

```JSON
{
  "landing": {
    "manifestLocation": "/apps/landing/fed-mods.json",
    "defaultDocumentTitle": "Hybrid Cloud Console Home",
    "isFedramp": true,
    "modules": [
      {
        "id": "landing",
        "module": "./RootApp",
        "routes": [
          {
            "permissions": [{
                "method": "withEmail",
                "args": ["@foo.bar"]
            }],
            "pathname": "/"
          }
        ]
      }
    ]
  },
}
```
