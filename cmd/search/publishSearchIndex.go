package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type SearchEnv string
type Release string

const (
	Prod          SearchEnv = "prod"
	Stage         SearchEnv = "stage"
	Itless        SearchEnv = "itless"
	Stable        Release   = "stable"
	Beta          Release   = "beta"
	ssoPathname   string    = "/auth/realms/redhat-external/protocol/openid-connect/token"
	hydraPathname string    = "/hydra/rest/search/console/index"
)

func (se SearchEnv) IsValidEnv() error {
	switch se {
	case Prod, Stage:
		return nil
	}

	return fmt.Errorf("invalid environment. Expected one of %s, %s, got %s", Prod, Stage, se)
}

type EnvMap map[SearchEnv]string

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type ModuleIndexEntry struct {
	Icon            string        `json:"icon,omitempty"`
	Title           []string      `json:"title"`
	Bundle          []string      `json:"bundle"`
	BundleTitle     []string      `json:"bundleTitle"`
	AltTitle        []string      `json:"alt_title,omitempty"`
	Id              string        `json:"id"`
	Uri             string        `json:"uri"`
	SolrCommand     string        `json:"solrCommand"`
	ContentType     string        `json:"contentType"`
	ViewUri         string        `json:"view_uri"`
	RelativeUri     string        `json:"relative_uri"`
	PocDescriptionT string        `json:"poc_description_t"`
	Permissions     []interface{} `json:"permissions,omitempty"`
}

type LinkEntry struct {
	Id          string        `json:"id"`
	Title       string        `json:"title"`
	Href        string        `json:"href"`
	Description string        `json:"description"`
	AltTitle    []string      `json:"alt_title,omitempty"`
	Permissions []interface{} `json:"permissions,omitempty"`
}

func findFirstValidChildLink(routes []interface{}) LinkEntry {
	result := LinkEntry{}
	for _, r := range routes {
		route, ok := r.(map[string]interface{})
		nestedRoutes, nestedOk := route["routes"].([]interface{})
		href, hrefOk := route["href"].(string)
		if hrefOk {
			result.Href = href
		} else if ok && route["expandable"] == true && nestedOk {
			// deeply nested item
			result = findFirstValidChildLink(nestedRoutes)
		}

		// exit if result was found
		if len(result.Href) > 0 {
			break
		}
	}

	return result
}

func convertAltTitles(jsonEntry interface{}) []string {
	altTitlesInterface, ok := jsonEntry.([]interface{})
	if !ok {
		// Cannot convert all title to array. Is probably empty.
		return []string{}
	}
	var altTitles []string
	for _, v := range altTitlesInterface {
		altTitles = append(altTitles, v.(string))
	}
	return altTitles
}

func parseLinkEntry(item map[string]interface{}) (LinkEntry, bool) {
	id, idOk := item["id"].(string)
	if !idOk || len(id) == 0 {
		return LinkEntry{}, false
	}

	title, titleOk := item["title"].(string)
	if !titleOk || len(title) == 0 {
		return LinkEntry{}, false
	}

	if item["expandable"] == true {
		childLink := findFirstValidChildLink(item["routes"].([]interface{}))
		item["href"] = childLink.Href
	}

	href, hrefOk := item["href"].(string)
	if !hrefOk || len(href) == 0 {
		return LinkEntry{}, false
	}

	var permissions []interface{}
	permissions = nil
	if item["permissions"] != nil {
		p, permissionsOk := item["permissions"].([]interface{})
		if !permissionsOk {
			fmt.Println("[Error]: permissions are not an array in link: ", id)
			return LinkEntry{}, false
		}
		permissions = p
	}

	return LinkEntry{
		Id:          id,
		Title:       title,
		Href:        href,
		Permissions: permissions,
	}, true
}

func flattenLinks(data interface{}, locator string) ([]LinkEntry, error) {
	flatData := []LinkEntry{}

	topLevel, ok := data.(map[string]interface{})
	// this is top section or a group item of nav file
	if ok && topLevel["navItems"] != nil {
		data, err := flattenLinks(topLevel["navItems"], fmt.Sprintf("%s.%s", locator, "navItems"))
		return append(flatData, data...), err
	}

	// argument came in as an array
	isArray, ok := data.([]interface{})
	if ok {
		for i, item := range isArray {
			items, err := flattenLinks(item, fmt.Sprintf("%s[%d]", locator, i))
			if err != nil {
				return []LinkEntry{}, err
			}
			flatData = append(flatData, items...)
		}

		return flatData, nil
	}

	routes, routesOk := topLevel["routes"].([]interface{})
	id, idOk := topLevel["id"].(string)
	// expandable item is a valid indexable item
	if topLevel["expandable"] == true && routesOk && idOk {
		// need to find a firs valid child route
		link := findFirstValidChildLink(routes)
		link.Id = id
		link.Title = topLevel["title"].(string)
		// Alternative titles are optional
		if topLevel["alt_title"] != nil {
			link.AltTitle = convertAltTitles(topLevel["alt_title"])
		}
		description, ok := topLevel["description"].(string)
		if ok {
			link.Description = description
		}
		flatData = append(flatData, link)
	}

	if topLevel["expandable"] == true && routesOk {
		for _, r := range routes {
			i, ok := r.(map[string]interface{})
			id, idOk := i["id"].(string)
			_, nestedRoutesOk := i["routes"].([]interface{})
			if ok && idOk {
				// all of these are required and type assertion can't fail
				link, linkOk := parseLinkEntry(i)
				if !linkOk {
					err := fmt.Errorf("[ERROR] Expandable: parsing link for href entry at %s", locator)
					return []LinkEntry{}, err
				}

				// Alternative titles are optional
				if i["alt_title"] != nil {
					link.AltTitle = convertAltTitles(i["alt_title"])
				}

				// description is optional
				description, ok := i["description"].(string)
				if ok {
					link.Description = description
				}
				flatData = append(flatData, link)
			} else if nestedRoutesOk {
				nestedItems, err := flattenLinks(r, fmt.Sprintf("%s.%s", locator, "routes"))
				if err != nil {
					return []LinkEntry{}, err
				}
				flatData = append(flatData, nestedItems...)
			} else {
				fmt.Printf("[WARN] Unable to convert link id %v to string. %v in file %s\n", id, i, locator)
			}
		}
		return flatData, nil
	}

	// this is directly a link
	item, ok := data.(map[string]interface{})
	if ok {
		href, ok := item["href"].(string)
		if ok && len(href) > 0 {
			link, linkOk := parseLinkEntry(item)
			if !linkOk {
				err := fmt.Errorf("[ERROR] parsing link for href entry at %s", locator)
				return []LinkEntry{}, err
			}

			// Alternative titles are optional
			if topLevel["alt_title"] != nil {
				link.AltTitle = convertAltTitles(topLevel["alt_title"])
			}

			// description is optional
			description, ok := item["description"].(string)
			if ok {
				link.Description = description
			}
			flatData = append(flatData, link)
		}
	}
	return flatData, nil
}

type groupLinkTemplate struct {
	Id      string        `json:"id"`
	IsGroup bool          `json:"isGroup"`
	Title   string        `json:"title"`
	Links   []interface{} `json:"links"`
}

type servicesTemplate struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Title       string `json:"title"`
	// Links can be []links or groupLinkTemplate
	Links []interface{} `json:"links"`
}

type ServiceLink struct {
	LinkEntry
	IsGroup  bool        `json:"-"`
	Links    []LinkEntry `json:"links,omitempty"`
	AltTitle []string    `json:"alt_title,omitempty"`
}

type ServiceEntry struct {
	servicesTemplate
	Links []ServiceLink `json:"links,omitempty"`
}

func findLinkById(id string, flatLinks []LinkEntry) (ServiceLink, bool) {
	var link ServiceLink
	for _, l := range flatLinks {
		if l.Id == id {
			link.LinkEntry = l
			link.AltTitle = l.AltTitle
			return link, true
		}
	}

	return link, false
}

func injectLinks(templateData []byte, flatLinks []LinkEntry) ([]ServiceEntry, error) {
	var templates []servicesTemplate
	var services []ServiceEntry
	err := json.Unmarshal(templateData, &templates)
	if err != nil {
		return services, err
	}

	for _, v := range templates {
		finalLinks := []ServiceLink{}
		serviceEntry := ServiceEntry{
			v,
			finalLinks,
		}
		for _, link := range v.Links {
			// is not a group links section
			stringLink, ok := link.(string)
			if ok {
				entry, found := findLinkById(stringLink, flatLinks)
				if found {
					finalLinks = append(finalLinks, entry)
				}
			}
			// is a group link
			g, ok := link.(map[string]interface{})
			gStr, err := json.Marshal(g)
			if err == nil {
				var group groupLinkTemplate
				err := json.Unmarshal(gStr, &group)
				if err == nil {
					if ok {
						for _, stringLink := range group.Links {
							var castLink string
							castLink, ok = stringLink.(string)
							var found bool
							var entry ServiceLink
							if ok {
								entry, found = findLinkById(castLink, flatLinks)
							}
							/**
							* Else branch is now handled because the link is "artificial" and does not exist in navigation.
							* If a link is not in the navigation, it is not from the link registry ad has to be added manually.
							 */
							if found {
								finalLinks = append(finalLinks, entry)
							} else if !found && !ok {
								linkMap, ok := stringLink.(map[string]interface{})
								if !ok {
									continue
								}

								artificialLink := ServiceLink{
									LinkEntry: LinkEntry{
										Title: linkMap["title"].(string),
										Href:  linkMap["href"].(string),
									},
								}

								description, ok := linkMap["description"].(string)
								artificialLink.AltTitle = convertAltTitles(linkMap["alt_title"])

								if ok {
									artificialLink.Description = description
								}
								finalLinks = append(finalLinks, artificialLink)

							}
						}
					}

				}

				// custom static entry
				custom, customOk := g["custom"].(bool)
				if customOk && custom {
					var customLink ServiceLink
					err := json.Unmarshal(gStr, &customLink)
					if err == nil {
						finalLinks = append(finalLinks, customLink)
					}
				}
			}
		}
		serviceEntry.Links = finalLinks
		services = append(services, serviceEntry)
	}

	return services, nil
}

func flattenIndexBase(indexBase []ServiceEntry, env SearchEnv) ([]ModuleIndexEntry, error) {
	hccOrigins := EnvMap{
		Prod:   "https://console.redhat.com",
		Stage:  "https://console.stage.redhat.com",
		Itless: "https://console.openshiftusgov.com",
	}
	bundleMapping := map[string]string{
		"application-services": "Application Services",
		"openshift":            "OpenShift",
		"ansible":              "Ansible",
		"insights":             "RHEL",
		"edge":                 "Edge Management",
		"settings":             "Settings",
		"landing":              "Home",
		"allservices":          "Home",
		"iam":                  "Identity & Access Management",
		"internal":             "Internal",
		"containers":           "Containers",
		"quay":                 "Quay.io",
	}
	var flatLinks []ModuleIndexEntry
	for _, s := range indexBase {
		for _, e := range s.Links {
			bundle := strings.Split(e.Href, "/")[1]
			newLink := ModuleIndexEntry{
				Icon:            s.Icon,
				PocDescriptionT: e.Description,
				Title:           []string{e.Title},
				Bundle:          []string{bundle},
				BundleTitle:     []string{bundleMapping[bundle]},
				Id:              fmt.Sprintf("hcc-module-%s-%s", e.Href, e.Id),
				Uri:             fmt.Sprintf("%s%s", hccOrigins[env], e.Href),
				SolrCommand:     "index",
				ContentType:     "moduleDefinition",
				ViewUri:         fmt.Sprintf("%s%s", hccOrigins[env], e.Href),
				RelativeUri:     e.Href,
				AltTitle:        e.AltTitle,
				Permissions:     e.Permissions,
			}
			flatLinks = append(flatLinks, newLink)
		}
	}
	return flatLinks, nil
}

// create search index compatible documents array
func constructIndex(env SearchEnv, release Release) ([]ModuleIndexEntry, error) {
	// get services template file
	stageContent, err := ioutil.ReadFile(fmt.Sprintf("static/%s/%s/services/services.json", release, env))
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	// get static service template only for search index
	// TODO: Add releases for static services
	staticContent, err := ioutil.ReadFile("cmd/search/static-services-entries.json")
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	// get all environment navigation files paths request to fill in template file
	stageNavFiles, err := filepath.Glob(fmt.Sprintf("static/%s/%s/navigation/*-navigation.json", release, env))
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	flatLinks := []LinkEntry{}
	for _, file := range stageNavFiles {
		var navItemData interface{}
		navFile, err := ioutil.ReadFile(file)
		if !strings.Contains(file, "landing") {
			if err != nil {
				return []ModuleIndexEntry{}, err
			}
			err = json.Unmarshal(navFile, &navItemData)
			if err != nil {
				return []ModuleIndexEntry{}, err
			}

			flatData, err := flattenLinks(navItemData, file)
			if err != nil {
				return []ModuleIndexEntry{}, err
			}
			// add group ID to link id
			fragments := strings.Split(strings.Split(file, "-navigation.json")[0], "/")
			navGroupId := fragments[len(fragments)-1]
			for index, link := range flatData {
				link.Id = fmt.Sprintf("%s.%s", navGroupId, link.Id)
				flatData[index] = link
			}
			flatLinks = append(flatLinks, flatData...)
		}
	}
	indexBase, err := injectLinks(stageContent, flatLinks)
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	staticBase, err := injectLinks(staticContent, flatLinks)
	if err != nil {
		return []ModuleIndexEntry{}, err
	}
	envIndex, err := flattenIndexBase(append(indexBase, staticBase...), env)

	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	return envIndex, nil
}

func getEnvToken(secret string, host string) (string, error) {
	data := url.Values{}
	// set payload data
	data.Set("client_id", "CRC-search-indexing")
	data.Set("grant_type", "client_credentials")
	data.Set("scope", "email profile openid")
	data.Set("client_secret", secret)

	// create request and encode data
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", host, ssoPathname), strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	// add request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	// fire request
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	// parse body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(body)
	// handle non 200 response
	if res.StatusCode >= 400 {
		return bodyString, errors.New(bodyString)
	}
	defer res.Body.Close()

	// retrieve access token
	var respJson TokenResponse
	err = json.Unmarshal([]byte(bodyString), &respJson)
	if err != nil {
		return "", err
	}

	return respJson.AccessToken, nil
}

type UploadPayload struct {
	DataSource string             `json:"dataSource"`
	Documents  []ModuleIndexEntry `json:"documents"`
}

func uploadIndex(token string, index []ModuleIndexEntry, host string) error {
	// fmt.Println(index)
	// return nil
	payload := UploadPayload{
		DataSource: "console",
		Documents:  index,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", host, hydraPathname), bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   "squid.corp.redhat.com:3128",
			}),
		},
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	// parse body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	bodyString := string(body)
	// handle non 200 response
	if res.StatusCode >= 400 {
		return errors.New(bodyString)
	}
	defer res.Body.Close()

	return nil
}

func deployIndex(env SearchEnv, envSecret string, ssoHost string, hydraHost string) error {
	err := env.IsValidEnv()
	if err != nil {
		return err
	}
	token, err := getEnvToken(envSecret, ssoHost)
	if err != nil {
		return err
	}
	index, err := constructIndex(env, "stable")
	if err != nil {
		return err
	}

	err = uploadIndex(token, index, hydraHost)
	if err != nil {
		return err
	}

	return nil
}

func handleErrors(errors []error, dryRun bool) {
	if len(errors) == 0 {
		fmt.Println("Search index published successfully")
	} else {
		for _, e := range errors {
			fmt.Println(e)
		}
		fmt.Println("Search index publishing failed. See above errors.")
		if dryRun {
			os.Exit(1)
		}
	}
}

func main() {
	// load env variables
	godotenv.Load()
	fmt.Println("Publishing search index")
	secrets := EnvMap{
		Prod:  os.Getenv("SEARCH_CLIENT_SECRET_PROD"),
		Stage: os.Getenv("SEARCH_CLIENT_SECRET_STAGE"),
	}

	ssoHosts := EnvMap{
		Prod:  "https://sso.redhat.com",
		Stage: "https://sso.stage.redhat.com",
	}

	hydraHost := EnvMap{
		Prod:  "https://access.redhat.com",
		Stage: "https://access.stage.redhat.com",
	}

	dryRun, _ := strconv.ParseBool(os.Getenv("SEARCH_INDEX_DRY_RUN"))
	writeIndex, _ := strconv.ParseBool(os.Getenv("SEARCH_INDEX_WRITE"))

	fmt.Println("Write index:", writeIndex)
	errors := []error{}

	if writeIndex {
		cwd, err := filepath.Abs(".")
		if err != nil {
			fmt.Println("Failed to get current working directory")
			errors = append(errors, err)
			handleErrors(errors, dryRun)
			return
		}
		writeEnvs := []SearchEnv{Prod, Stage, Itless}
		writeReleases := []Release{Stable, Beta}
		for _, env := range writeEnvs {
			for _, release := range writeReleases {
				searchIndex, err := constructIndex(env, release)
				if err != nil {
					fmt.Println("Failed to construct search index for", env, release)
					errors = append(errors, err)
				} else {
					dirname := fmt.Sprintf("%s/static/%s/%s/search", cwd, release, env)
					fileName := fmt.Sprintf("%s/search-index.json", dirname)
					err := os.MkdirAll(dirname, os.ModePerm)
					if err != nil {
						fmt.Println("Failed to create directory", dirname)
						errors = append(errors, err)
					} else {
						j, err := json.Marshal(searchIndex)
						if err != nil {
							fmt.Println("Failed to marshal search index")
							errors = append(errors, err)
						}
						err = os.WriteFile(fileName, j, 0644)
						if err != nil {
							fmt.Println("Failed to write search index to", fileName)
							errors = append(errors, err)
						}
					}

				}

			}
		}
		handleErrors(errors, dryRun)
		return
	}

	for _, env := range []SearchEnv{Stage, Prod} {
		var err error
		if dryRun {
			fmt.Println("Attempt dry run search index for", env, "environment.")
			_, err = constructIndex(env, "stable")
		} else {
			fmt.Println("Attempt to publish search index for", env, "environment.")
			err = deployIndex(env, secrets[env], ssoHosts[env], hydraHost[env])
		}
		if err != nil {
			fmt.Println("[ERROR] Failed to deploy search index for", env, "environment.")
			fmt.Println(err)
			errors = append(errors, err)
		}
	}

	handleErrors(errors, dryRun)
}
