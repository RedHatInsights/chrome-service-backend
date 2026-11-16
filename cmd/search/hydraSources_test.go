package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveURI(t *testing.T) {
	assert.Equal(t, "https://console.stage.redhat.com/insights/foo", resolveURI("/insights/foo", Stage))
	assert.Equal(t, "https://www.redhat.com/trial", resolveURI("https://www.redhat.com/trial", Stage))
}

func TestBundleFromHref(t *testing.T) {
	assert.Equal(t, "insights", bundleFromHref("/insights/advisor"))
	assert.Equal(t, "", bundleFromHref("/?quickstart=foo"))
	assert.Equal(t, "", bundleFromHref("https://www.redhat.com/x"))
}

func TestLoadFeoSearchIndex(t *testing.T) {
	path := filepath.Join(t.TempDir(), "feo.json")
	content := `[
		{"id": "acs", "href": "/openshift/acs/overview", "title": "ACS", "description": "desc", "alt_title": ["acs"]},
		{"id": "broken", "title": "No href"}
	]`
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))

	docs, err := loadFeoSearchIndex(path, Stage)
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
	assert.Equal(t, "hcc-feo-acs", docs[0].Id)
	assert.Equal(t, []string{"openshift"}, docs[0].Bundle)
	assert.Equal(t, []string{"OpenShift"}, docs[0].BundleTitle)
	assert.Equal(t, "https://console.stage.redhat.com/openshift/acs/overview", docs[0].Uri)
	assert.Equal(t, moduleContentType, docs[0].ContentType)
}

func TestLoadQuickstarts(t *testing.T) {
	dir := t.TempDir()
	write := func(rel string, content string) {
		p := filepath.Join(dir, rel)
		assert.NoError(t, os.MkdirAll(filepath.Dir(p), 0755))
		assert.NoError(t, os.WriteFile(p, []byte(content), 0644))
	}
	write("interactive/interactive.yml", `
metadata:
  name: interactive
spec:
  displayName: Interactive QS
  description: Do things
  tasks:
    - title: step
`)
	write("interactive/metadata.yml", `
kind: QuickStarts
name: interactive
tags:
  - kind: bundle
    value: insights
`)
	write("external/external.yml", `
kind: QuickStarts
metadata:
  name: external
  externalDocumentation: true
spec:
  displayName: External QS
  description: Read docs
  link:
    href: https://docs.redhat.com/external
`)
	write("other.yml", "foo: bar\n")

	docs, err := loadQuickstarts(dir, Prod)
	assert.NoError(t, err)
	assert.Len(t, docs, 2)

	byId := map[string]ModuleIndexEntry{}
	for _, d := range docs {
		byId[d.Id] = d
	}
	interactive := byId["hcc-quickstart-interactive"]
	assert.Equal(t, "https://console.redhat.com/?quickstart=interactive", interactive.Uri)
	assert.Equal(t, []string{"insights"}, interactive.Bundle)
	assert.Equal(t, []string{"RHEL"}, interactive.BundleTitle)
	assert.Equal(t, quickstartContentType, interactive.ContentType)

	external := byId["hcc-quickstart-external"]
	assert.Equal(t, "https://docs.redhat.com/external", external.Uri)
	assert.Equal(t, []string{""}, external.Bundle)
}

func TestMergeIndexes(t *testing.T) {
	feo := []ModuleIndexEntry{{Id: "feo-a", Uri: "u1", Title: []string{"A"}}}
	modules := []ModuleIndexEntry{
		{Id: "hcc-module-a", Uri: "u1", Title: []string{"A legacy"}, AltTitle: []string{"a", "alias"}, PocDescriptionT: "desc", Icon: "Icon"},
		{Id: "mod-b", Uri: "u2", Title: []string{"B"}},
		{Id: "mod-b", Uri: "u3", Title: []string{"B dup id"}},
	}

	merged := mergeIndexes(feo, modules)
	assert.Len(t, merged, 2)
	// FEO content wins, but the already published module id is kept so Hydra overwrites it
	assert.Equal(t, "hcc-module-a", merged[0].Id)
	assert.Equal(t, []string{"A"}, merged[0].Title)
	assert.Equal(t, []string{"A legacy", "alias"}, merged[0].AltTitle)
	assert.Equal(t, "desc", merged[0].PocDescriptionT)
	assert.Equal(t, "Icon", merged[0].Icon)
	assert.Equal(t, "mod-b", merged[1].Id)
}
