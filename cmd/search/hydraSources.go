package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	moduleContentType     = "moduleDefinition"
	quickstartContentType = "quickstart"
	legacyModuleIdPrefix  = "hcc-module-"
)

var hccOrigins = EnvMap{
	Prod:   "https://console.redhat.com",
	Stage:  "https://console.stage.redhat.com",
	Itless: "https://console.openshiftusgov.com",
}

var bundleMapping = map[string]string{
	"application-services": "Application Services",
	"openshift":            "OpenShift",
	"ansible":              "Ansible",
	"insights":             "RHEL",
	"settings":             "Settings",
	"landing":              "Home",
	"allservices":          "Home",
	"iam":                  "Identity & Access Management",
	"internal":             "Internal",
	"containers":           "Containers",
	"quay":                 "Quay.io",
	"subscriptions":        "Subscriptions",
}

func isAbsoluteURL(href string) bool {
	return strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://")
}

// resolveURI returns the full URL for an href. Console-relative hrefs are prefixed
// with the environment origin; absolute (external) hrefs are returned untouched.
func resolveURI(href string, env SearchEnv) string {
	if isAbsoluteURL(href) {
		return href
	}
	return hccOrigins[env] + href
}

// bundleFromHref returns the first path segment of a console-relative href.
func bundleFromHref(href string) string {
	if isAbsoluteURL(href) {
		return ""
	}
	parts := strings.Split(strings.TrimPrefix(href, "/"), "/")
	return strings.SplitN(parts[0], "?", 2)[0]
}

// FeoSearchEntry is a single entry of the frontend-operator generated search index
// (the FEO_SEARCH_INDEX config map value / static/search-index-generated.json).
type FeoSearchEntry struct {
	Id          string   `json:"id"`
	Href        string   `json:"href"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	AltTitle    []string `json:"alt_title"`
	FrontendRef string   `json:"frontendRef"`
}

func loadFeoSearchIndex(path string, env SearchEnv) ([]ModuleIndexEntry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading FEO search index %s: %w", path, err)
	}
	var entries []FeoSearchEntry
	if err := json.Unmarshal(content, &entries); err != nil {
		return nil, fmt.Errorf("parsing FEO search index %s: %w", path, err)
	}

	docs := make([]ModuleIndexEntry, 0, len(entries))
	for _, e := range entries {
		if e.Id == "" || e.Title == "" || e.Href == "" {
			fmt.Printf("[WARN] Skipping FEO search entry with missing id/title/href: %+v\n", e)
			continue
		}
		bundle := bundleFromHref(e.Href)
		docs = append(docs, ModuleIndexEntry{
			Title:           []string{e.Title},
			Bundle:          []string{bundle},
			BundleTitle:     []string{bundleMapping[bundle]},
			AltTitle:        e.AltTitle,
			Id:              fmt.Sprintf("hcc-feo-%s", e.Id),
			Uri:             resolveURI(e.Href, env),
			SolrCommand:     "index",
			ContentType:     moduleContentType,
			ViewUri:         resolveURI(e.Href, env),
			RelativeUri:     e.Href,
			PocDescriptionT: e.Description,
		})
	}
	return docs, nil
}

type quickstartFile struct {
	Metadata struct {
		Name                  string `yaml:"name"`
		ExternalDocumentation bool   `yaml:"externalDocumentation"`
	} `yaml:"metadata"`
	Spec struct {
		DisplayName string `yaml:"displayName"`
		Description string `yaml:"description"`
		Type        struct {
			Text string `yaml:"text"`
		} `yaml:"type"`
		Link struct {
			Href string `yaml:"href"`
		} `yaml:"link"`
		Tasks []interface{} `yaml:"tasks"`
	} `yaml:"spec"`
}

type quickstartMetadata struct {
	Name string `yaml:"name"`
	Tags []struct {
		Kind  string `yaml:"kind"`
		Value string `yaml:"value"`
	} `yaml:"tags"`
}

// loadQuickstarts walks a checkout of the quickstarts repository content directory
// (github.com/RedHatInsights/quickstarts, docs/quickstarts) and converts every
// quickstart definition into a search document.
func loadQuickstarts(dir string, env SearchEnv) ([]ModuleIndexEntry, error) {
	// metadata.yml files hold the bundle tags, keyed by quickstart name
	bundlesByName := map[string][]string{}
	var definitions []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".yml" && ext != ".yaml" {
			return nil
		}
		if d.Name() == "metadata.yml" || d.Name() == "metadata.yaml" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var meta quickstartMetadata
			if err := yaml.Unmarshal(content, &meta); err != nil {
				fmt.Printf("[WARN] Unable to parse quickstart metadata %s: %v\n", path, err)
				return nil
			}
			for _, t := range meta.Tags {
				if t.Kind == "bundle" && t.Value != "" {
					bundlesByName[meta.Name] = append(bundlesByName[meta.Name], strings.ToLower(t.Value))
				}
			}
			return nil
		}
		definitions = append(definitions, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking quickstarts directory %s: %w", dir, err)
	}

	sort.Strings(definitions)
	docs := []ModuleIndexEntry{}
	for _, path := range definitions {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var qs quickstartFile
		if err := yaml.Unmarshal(content, &qs); err != nil {
			fmt.Printf("[WARN] Unable to parse quickstart %s: %v\n", path, err)
			continue
		}
		name := qs.Metadata.Name
		if name == "" || qs.Spec.DisplayName == "" {
			// not a quickstart definition
			continue
		}

		var uri, relativeUri string
		if qs.Metadata.ExternalDocumentation && isAbsoluteURL(qs.Spec.Link.Href) {
			// documentation only quickstart, points to external docs
			uri = qs.Spec.Link.Href
			relativeUri = qs.Spec.Link.Href
		} else {
			// in-console quickstart; chrome opens the quickstart drawer from the query param on any page
			relativeUri = fmt.Sprintf("/?quickstart=%s", name)
			uri = resolveURI(relativeUri, env)
		}

		bundles := bundlesByName[name]
		if len(bundles) == 0 {
			bundles = []string{""}
		}
		bundleTitles := make([]string, 0, len(bundles))
		for _, b := range bundles {
			bundleTitles = append(bundleTitles, bundleMapping[b])
		}

		docs = append(docs, ModuleIndexEntry{
			Title:           []string{qs.Spec.DisplayName},
			Bundle:          bundles,
			BundleTitle:     bundleTitles,
			Id:              fmt.Sprintf("hcc-quickstart-%s", name),
			Uri:             uri,
			SolrCommand:     "index",
			ContentType:     quickstartContentType,
			ViewUri:         uri,
			RelativeUri:     relativeUri,
			PocDescriptionT: strings.TrimSpace(qs.Spec.Description),
		})
	}
	return docs, nil
}

// mergeIndexes concatenates search documents from multiple sources. Documents pointing
// to the same URI are collapsed into one; the earlier source wins, alternative titles
// are merged and a missing description/icon is backfilled from later sources. When a
// collapsed document has a legacy module id, that id is kept (see below).
func mergeIndexes(sources ...[]ModuleIndexEntry) []ModuleIndexEntry {
	merged := []ModuleIndexEntry{}
	byUri := map[string]int{}
	seenIds := map[string]bool{}
	for _, source := range sources {
		for _, doc := range source {
			if i, ok := byUri[doc.Uri]; ok {
				existing := &merged[i]
				// Hydra upserts by id and never deletes. Keep the id of the already published
				// module document so it is overwritten instead of duplicated.
				if strings.HasPrefix(doc.Id, legacyModuleIdPrefix) && !strings.HasPrefix(existing.Id, legacyModuleIdPrefix) && !seenIds[doc.Id] {
					delete(seenIds, existing.Id)
					existing.Id = doc.Id
					seenIds[doc.Id] = true
				}
				existing.AltTitle = mergeAltTitles(existing.AltTitle, append(doc.Title, doc.AltTitle...), existing.Title)
				if existing.PocDescriptionT == "" {
					existing.PocDescriptionT = doc.PocDescriptionT
				}
				if existing.Icon == "" {
					existing.Icon = doc.Icon
				}
				continue
			}
			if seenIds[doc.Id] {
				fmt.Printf("[WARN] Duplicate search document id %s with different URI, skipping %s\n", doc.Id, doc.Uri)
				continue
			}
			seenIds[doc.Id] = true
			byUri[doc.Uri] = len(merged)
			merged = append(merged, doc)
		}
	}
	return merged
}

func mergeAltTitles(existing []string, extra []string, titles []string) []string {
	seen := map[string]bool{}
	for _, t := range titles {
		seen[strings.ToLower(t)] = true
	}
	result := []string{}
	for _, t := range append(append([]string{}, existing...), extra...) {
		key := strings.ToLower(t)
		if t == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, t)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
