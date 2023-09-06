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

### main.yml changes

The `main.yml` files is now used purely for the API frontend config. That is the last part of the platform using this file. Soon the contents will be migrated to the application manifest file.

The `main.yml` files can be found in the [`static`](https://github.com/RedHatInsights/chrome-service-backend/tree/main/static)

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

### Proxy route is different

```js
routes: {
  "/api/chrome-service/v1/static": {
    host: "http://localhost:<port>"
  }
}
```

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

## New feature outside of legacy cloud services config

### Third navigation level

Left navigation now supports extra nesting level. To add additional level to your navigation, add expandable navigation partial instead of leaf link to a nav array in a navigation file:

```js
// Navigation file of a bundle
{
  "id": "navSection",
  "title": "My nav",
  "navItems": [
    {
      // regular level of nesting
      "title": "Normal expandable item",
      "expandable": true,
      "routes": [
        {
          // extra nested links
          "title": "Deeply expandable item",
          "expandable": true,
          // regular routes definition
          "routes": [
            {
              "id": "linkId",
              "appId": "uiModuleId",
              "title": "Deep link",
              "href": "/path/to/screen",
              "description": "Description for all services"
            },
          ]
        }
      ]
    }
  ]
}
  
```

# All services dropdown and page<a name="allservices"></>

There is a new schema that controls which items will be display in the all services dropdown and pages.

The data is stored in `service.json` files within `static` directory. Each environment has its own definition file. For example, for `stage-beta` environment, the file is located at [/static/beta/stage/services/services.json](https://github.com/RedHatInsights/chrome-service-backend/blob/main/static/beta/stage/services/services.json)

**All the tiles (links) in the all services dropdown (page), must exist somewhere in the navigation files**. The tiles (links) are generated from these files. If no reference is found in the navigation files, the tile (link) will not show in the UI.

**Changes to the navigation files can affect the all services dropdown (page)**. Before you submit the changes for review, make sure the all services will be in expected state.

Be aware that changing the link IDs in navigation file and not reflecting the change in respective `services.json` will remove the tile from the all services dropdown (page).

## Link requirements

In order for a link to be eligible to display in the all services dropdown it must have following attributes inside the <bundle>-navigation.json file:

- `id`: a bundle unique id of a **link**
- `description`: short description of the page the link leads to
- `href`: pathname to be used

### Using expandable section as a link

If your leaf links (clickable links in left navigation) are not descriptive enough for the services dropdown or page (for example Overview or Systems), the expandable section can be used. **The HREF attribute will be picked from a first valid link of expandable section.**

In order to make expandable section eligible for usage in the all services dropdown/page, do the following:

1. add unique id to the expandable section definition
2. add short description of the section
3. follow the [Inserting link into services.json](#insertlink)

Example changes:

```diff
diff --git a/static/stable/stage/navigation/settings-navigation.json b/static/stable/stage/navigation/settings-navigation.json
index 30bb92f..16155ba 100644
--- a/static/stable/stage/navigation/settings-navigation.json
+++ b/static/stable/stage/navigation/settings-navigation.json
@@ -23,6 +23,8 @@
         {
             "title": "Notifications",
             "expandable": true,
+            "id": "foobar",
+            "description": "Some nice description....",
             "routes": [
                 {
                     "id": "openshift",
diff --git a/static/stable/stage/services/services.json b/static/stable/stage/services/services.json
index a1cfa81..f3e1037 100644
--- a/static/stable/stage/services/services.json
+++ b/static/stable/stage/services/services.json
@@ -6,7 +6,8 @@
     "title": "Application services",
     "links": [
       "application-services.apiManagement",
-      "application-services.serviceAccounts"
+      "application-services.serviceAccounts",
+      "settings.foobar"
     ]
   },
   {

```

### Using bundle title as a link

If your bundle has a root page with non descriptive title (like Overview for example), or your application is the entire bundle (like Quay), bundle titles can be used for all services tiles/links.

**The HREF attribute will be picked from a first valid link of the bundle.**

In order to make expandable section eligible for usage in the all services dropdown/page, do the following:

1. add short description of the section
2. follow the [Inserting link into services.json](#insertlink)

Example changes:

```diff
diff --git a/static/stable/stage/navigation/settings-navigation.json b/static/stable/stage/navigation/settings-navigation.json
index 30bb92f..9540df5 100644
--- a/static/stable/stage/navigation/settings-navigation.json
+++ b/static/stable/stage/navigation/settings-navigation.json
@@ -1,6 +1,7 @@
 {
     "id": "settings",
     "title": "Settings",
+    "description": "Some nice description...",
     "navItems": [
         {
             "id": "sources",
diff --git a/static/stable/stage/services/services.json b/static/stable/stage/services/services.json
index a1cfa81..98167dd 100644
--- a/static/stable/stage/services/services.json
+++ b/static/stable/stage/services/services.json
@@ -6,7 +6,8 @@
     "title": "Application services",
     "links": [
       "application-services.apiManagement",
-      "application-services.serviceAccounts"
+      "application-services.serviceAccounts",
+      "settings.settings"
     ]
   },
   {

```

## Inserting link into services.json<a name="insertlink"></a>

To add a link to `services.json` follow these steps:

1. Find relevant `services.json` based on environment.
2. Find relevant section (item in the top level array). 
3. (optional) Find a relevant group within a section.
4. Add new link ID into the `links` array. The id must be in following format `<navigation file prefix>.<link id>`

> Note: the navigation file prefix means the a part of the navigation file name before the `-navigation.json` string. For example prefix for `ansible-navigation.json` is `ansible`. In this case, valid id would be `ansible.linkId`.

## Inserting new section

If your link does not fit into any existing section and you need a new one, follow these steps.

1. Make sure the section is both UX and PM approved.
2. Find relevant `services.json` based on environment.
3. Based on your title, insert a new empty at a correct place inside the services array. The order is **sorted alphabetically** based on the title.
4. Add all required data into the object (example bellow).
5. (optional) Segment your section into groups (example bellow).
6. Add all required links into the `links` array.


**Basic section object**
```ts
{
  id: string; // unique section id
  title: string; // title of a section
  icon: "CloudUploadAltIcon" | "AutomationIcon" | "BoxesIcon" | "DatabaseIcon" | "RocketIcon" | "UsersIcon" | "InfrastructureIcon" | "BellIcon" | "ChartLineIcon" | "CloudSecurityIcon" | "CreditCardIcon" | "CogIcon" | "ShoppingCartIcon"; // icon representing section
  description: string; // short section description
  links: string[] // list of links
}
```

If you require a new icon, contact the platform experience team. Only PF svg icons are available for use!

**Section with grouped links**

If you require links to be divided into groups, yse the same for as for basic section with different links array format.

```ts
{
  links: {
    id: string // group id
    isGroup: true // mark item as group
    title: string // group title
    links: string[] // array of link ids
  }[]
}
```

**Basic section links and grouped section links can't be mixed!**
