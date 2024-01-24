# Search index

## Search index source

The search index has two sources:

1. The `services.json` template and by extension the navigation files
2. The `cmd/search/static-services-entries.json` for entries not defined in `services.json`

## Adding entries to search index

### Entry is in `services.json`

If a link is referenced in `services.json`, it will be included in the search index. No action is needed to add such entry into the search index. Any change in the navigation file will be reflected in the search index.

### Entry is not in `services.json`

If the required entry is in navigation file, but was simply not chosen to show up in the `services.json`, it can be references in the `cmd/search/static-services-entries.json`. This is a statically defined equivalent to live `services.json` used only for search indexing purposes. Only a search entry can be added into this file. 

There are two possible paths:

#### Entry has a link in some navigation file

Add the link reference to the `static-services-entries.json`. There a few defined sections purely for organization purposes. If your link does not fit into any existing section, create a new one. Then add the link reference to the `links` array. To learn how to make a link eligible to be references, read this [documentation](https://github.com/RedHatInsights/chrome-service-backend/blob/main/docs/cloud-services-config.md#all-services-dropdown-and-page).

#### Entry does not have a link and is unique to the search index

If the new entry is not present in any navigation files, a new entry has to be created. Follow these steps:

1. In the `static-services-entries.json` file pick an existing section or create a new one
2. Create an object definition for the new entry

This is an example of such synthetic search entry:

```json
{
  "custom": true,
  "id": "OPENSHIFT.cluster.create.aws",
  "href": "/openshift/create",
  "title": "Red Hat OpenShift Service on AWS",
  "alt_title": ["ROSA", "AWS", "OpenShift on AWS"],
  "description": "Foo bar description"
},
```

The final custom section can look similar to this:

```json
  {
    "id": "create-clusters",
    "links": [
      {
      "custom": true,
      "id": "OPENSHIFT.cluster.create.aws",
      "href": "/openshift/create",
      "title": "Red Hat OpenShift Service on AWS",
      "alt_title": ["ROSA", "AWS", "OpenShift on AWS"],
      "description": "Foo bar description"
      },
      {
      "custom": true,
      "id": "OPENSHIFT.cluster.create.dedicated",
      "href": "/openshift/create",
      "title": "Red Hat OpenShift Dedicated",
      "alt_title": ["OSD","Dedicated", "GCP", "AWS", "OpenShift on AWS", "OpenShift on GCP"] 
      },
      "openshift.some-link-id"
    ]
  }
```

## Publishing search index

To publish the search index follow these steps:

1. Make sure the correct files were changed. If you are changing navigation files, make sure to edit both stage and production files.
2. Make sure you are on RH wired network or on VPN
3. Run the `make publish-search-index` command from the project root
