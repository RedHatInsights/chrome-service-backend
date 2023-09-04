# Custom service SSO Scope

By default, hybrid cloud console requires only a base set of SSO scopes. If your service requires additional Keycloak scopes, you can instruct the chroming service to attempt user re-authentication to update the user session token.

If the chroming service finds that a UI module is requires a scope and the scope was not request in previous authentication request it will try to re-authenticate the user. **This does not mean the user will have to type credentials again**. The behavior of the re-auth depends on the account state and the scope requirements. It can be a silent which will result in a browser refresh, or a one time additional registration form might appear.

## Adding additional SSO Scopes:

To trigger re-authentication, the chroming service must know that specific UI module requires it. This can be done by adding the `ssoScopes` attribute to a module configuration in the `fed-modules.json` file. Follow these steps:

1. Find `fed-mods.json` for a relevant environment (prod/stage/stable/preview) file(s).
2. Find relevant UI module based on the module key.
3. Add the new scope(s) id to the `config.ssoScopes` array. If the array does not exist create it.

A sample changes can look like this:

```diff
diff --git a/static/beta/prod/modules/fed-modules.json b/static/beta/prod/modules/fed-modules.json
index bb8aab0..b716df6 100644
--- a/static/beta/prod/modules/fed-modules.json
+++ b/static/beta/prod/modules/fed-modules.json
@@ -162,7 +162,12 @@
                     }
                 ]
             }
-        ]
+        ],
+        "config": {
+            "ssoScopes": [
+                "rhfull"
+            ]
+        }
     },
     "openshift": {
         "manifestLocation": "/apps/openshift/fed-mods.json",

```
