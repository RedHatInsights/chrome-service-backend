package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServiceItemLink_StringAndGroup(t *testing.T) {
	storage := map[string]interface{}{
		"foo.bar": map[string]interface{}{"id": "bar", "title": "Bar"},
	}

	// resolves string id
	res := parseServiceItemLink("foo.bar", &storage)
	assert.Equal(t, storage["foo.bar"], res)

	// resolves group links and preserves direct maps
	group := map[string]interface{}{
		"isGroup": true,
		"links":   []interface{}{"foo.bar", map[string]interface{}{"id": "baz"}},
	}
	gres := parseServiceItemLink(group, &storage)
	links, ok := gres["links"].([]map[string]interface{})
	assert.True(t, ok, "expected group links to be []map[string]interface{}")
	assert.Len(t, links, 2)
	assert.Equal(t, "bar", links[0]["id"])
	assert.Equal(t, "baz", links[1]["id"])

	// invalid type panics
	assert.Panics(t, func() { _ = parseServiceItemLink(123, &storage) })
}

func TestParseServiceItem(t *testing.T) {
	storage := map[string]interface{}{
		"foo.leaf": map[string]interface{}{"id": "leaf", "title": "Leaf"},
	}
	tb := templateBase{
		Id:          "svc1",
		Icon:        "icon",
		Title:       "Service 1",
		Description: "Desc",
		Links: []interface{}{
			"foo.leaf",
			map[string]interface{}{"isGroup": true, "links": []interface{}{"foo.leaf"}},
		},
	}

	si := parseServiceItem(tb, &storage)
	assert.Equal(t, "svc1", si["id"])
	assert.Equal(t, "icon", si["icon"])
	assert.Equal(t, "Service 1", si["title"])
	assert.Equal(t, "Desc", si["description"])
	links, ok := si["links"].([]interface{})
	assert.True(t, ok, "expected links to be []interface{}")
	assert.Len(t, links, 2)
	// first link is resolved map
	if l0, ok := links[0].(map[string]interface{}); !ok {
		t.Fatalf("unexpected first link type: %#v", links[0])
	} else {
		assert.Equal(t, "leaf", l0["id"])
	}
	// second link is a group with resolved inner links
	if g1, ok := links[1].(map[string]interface{}); !ok {
		t.Fatalf("unexpected second link type: %#v", links[1])
	} else {
		gLinks, ok := g1["links"].([]map[string]interface{})
		assert.True(t, ok, "expected group links to be []map[string]interface{}")
		assert.Len(t, gLinks, 1)
		assert.Equal(t, "leaf", gLinks[0]["id"])
	}
}

func TestParseNestedLinks(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"id": "a"},
		map[string]interface{}{"id": "b"},
	}
	out := parseNestedLinks(input)
	assert.Len(t, out, 2)
	assert.Equal(t, "a", out[0]["id"])
	assert.Equal(t, "b", out[1]["id"])

	// invalid element panics
	assert.Panics(t, func() { _ = parseNestedLinks([]interface{}{"oops"}) })
}

func TestFindFirstValidLeaf(t *testing.T) {
	// simple leaf
	items := []map[string]interface{}{{"id": "leaf1"}}
	leaf, found := findFirstValidLeaf(items)
	assert.True(t, found)
	assert.Equal(t, "leaf1", leaf["id"])

	// expandable with routes
	items = []map[string]interface{}{
		{"expandable": true, "routes": []interface{}{map[string]interface{}{"id": "r1"}}},
	}
	leaf, found = findFirstValidLeaf(items)
	assert.True(t, found)
	assert.Equal(t, "r1", leaf["id"])

	// nested navItems
	items = []map[string]interface{}{
		{"navItems": []interface{}{map[string]interface{}{"id": "inner"}}},
	}
	leaf, found = findFirstValidLeaf(items)
	assert.True(t, found)
	assert.Equal(t, "inner", leaf["id"])
}

func TestParseNavigationLinks(t *testing.T) {
	items := []map[string]interface{}{
		{"id": "parent", "title": "Parent", "description": "P", "navItems": []interface{}{map[string]interface{}{"id": "inner", "title": "Inner"}}},
		{"id": "exp", "title": "Exp", "description": "E", "expandable": true, "routes": []interface{}{map[string]interface{}{"id": "r"}}},
		{"id": "plain", "title": "Plain"},
	}
	flat := parseNavigationLinks(items)
	assert.Len(t, flat, 5)
	assert.Equal(t, "parent", flat[0]["id"])
	assert.Equal(t, "inner", flat[1]["id"])
	assert.Equal(t, "exp", flat[2]["id"])
	assert.Equal(t, "r", flat[3]["id"])
	assert.Equal(t, "plain", flat[4]["id"])
}

func TestGetLinksStorage(t *testing.T) {
	if got := getLinksStorage("stable", "stage"); got != &navLinksStorage.Stable.Stage {
		assert.Fail(t, "unexpected storage pointer for stable/stage")
	}
	if got := getLinksStorage("beta", "prod"); got != &navLinksStorage.Beta.Prod {
		assert.Fail(t, "unexpected storage pointer for beta/prod")
	}

	assert.Panics(t, func() { _ = getLinksStorage("stable", "unknown") })
	assert.Panics(t, func() { _ = getLinksStorage("unknown", "stage") })
}

func TestParseEnvironmentLinks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bundle-navigation.json")
	content := `{"navItems":[{"id":"leaf","title":"Leaf"}]}`
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))

	storage := map[string]interface{}{}
	parseEnvironmentLinks(path, &storage)
	_, ok := storage["bundle.leaf"]
	assert.True(t, ok, "expected storage to contain key bundle.leaf")

	// duplicate key should panic
	assert.Panics(t, func() { parseEnvironmentLinks(path, &storage) })
}

func TestParseEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "services.json")
	content := `[
		{"id":"svc1","icon":"i","title":"t","description":"d","links":["bundle.leaf", {"isGroup":true, "links":["bundle.leaf"]}]}
	]`
	assert.NoError(t, os.WriteFile(path, []byte(content), 0644))

	storage := map[string]interface{}{
		"bundle.leaf": map[string]interface{}{"id": "leaf", "title": "Leaf"},
	}

	out := parseEnvironment(path, &storage)
	assert.Len(t, out, 1)
	assert.Equal(t, "svc1", out[0]["id"])
	assert.Equal(t, "t", out[0]["title"])
	links, ok := out[0]["links"].([]interface{})
	assert.True(t, ok, "expected links to be []interface{}")
	assert.Len(t, links, 2)
}
