package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

type SearchEnv string

const (
	Prod          SearchEnv = "prod"
	Stage         SearchEnv = "stage"
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
	Icon            string   `json:"icon,omitempty"`
	Title           []string `json:"title"`
	Bundle          []string `json:"bundle"`
	BundleTitle     []string `json:"bundleTitle"`
	Id              string   `json:"id"`
	Uri             string   `json:"uri"`
	SolrCommand     string   `json:"solrCommand"`
	ContentType     string   `json:"contentType"`
	ViewUri         string   `json:"view_uri"`
	RelativeUri     string   `json:"relative_uri"`
	PocDescriptionT string   `json:"poc_description_t"`
}

type LinkEntry struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Href        string `json:"href"`
	Description string `json:"description"`
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

func flattenLinks(data interface{}) ([]LinkEntry, error) {
	flatData := []LinkEntry{}

	topLevel, ok := data.(map[string]interface{})
	// this is top section or a group item of nav file
	if ok && topLevel["navItems"] != nil {
		data, err := flattenLinks(topLevel["navItems"])
		return append(flatData, data...), err
	}

	// argument came in as an array
	isArray, ok := data.([]interface{})
	if ok {
		for _, item := range isArray {
			items, err := flattenLinks(item)
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
		description, ok := topLevel["description"].(string)
		if ok {
			link.Description = description
		}
		flatData = append(flatData, link)
	}

	if topLevel["expandable"] == true && routesOk {
		for _, r := range routes {
			i, ok := r.(map[string]interface{})
			if ok {
				// all of these are required and type assertion can't fail
				link := LinkEntry{
					Id:    i["id"].(string),
					Title: i["title"].(string),
					Href:  i["href"].(string),
				}

				// description is optional
				description, ok := i["description"].(string)
				if ok {
					link.Description = description
				}
				flatData = append(flatData, link)
			}
		}
		return flatData, nil
	}

	// this is directly a link
	item, ok := data.(map[string]interface{})
	if ok {
		href, ok := item["href"].(string)
		if ok && len(href) > 0 {
			link := LinkEntry{
				Id:    item["id"].(string),
				Title: item["title"].(string),
				Href:  item["href"].(string),
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
	Id      string   `json:"id"`
	IsGroup bool     `json:"isGroup"`
	Title   string   `json:"title"`
	Links   []string `json:"links"`
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
	IsGroup bool        `json:"-"`
	Links   []LinkEntry `json:"links,omitempty"`
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
							entry, found := findLinkById(stringLink, flatLinks)
							if found {
								finalLinks = append(finalLinks, entry)
							}
						}
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
		Prod:  "https://console.redhat.com",
		Stage: "https://console.stage.redhat.com",
	}
	bundleMapping := map[string]string{
		"application-services": "Application and Data Services",
		"openshift":            "OpenShift",
		"ansible":              "Ansible Automation Platform",
		"insights":             "Red Hat Insights",
		"edge":                 "Edge management",
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
				Id:              fmt.Sprintf("hcc-module-%s", e.Href),
				Uri:             fmt.Sprintf("%s%s", hccOrigins[env], e.Href),
				SolrCommand:     "index",
				ContentType:     "moduleDefinition",
				ViewUri:         fmt.Sprintf("%s%s", hccOrigins[env], e.Href),
				RelativeUri:     e.Href,
			}
			flatLinks = append(flatLinks, newLink)
		}
	}
	return flatLinks, nil
}

// create search index compatible documents array
func constructIndex(env SearchEnv) ([]ModuleIndexEntry, error) {
	// get services template file
	stageContent, err := ioutil.ReadFile(fmt.Sprintf("static/stable/%s/services/services.json", env))
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	// get static service template only for search index
	staticContent, err := ioutil.ReadFile("cmd/search/static-services-entries.json")
	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	// get all environment navigation files paths request to fill in template file
	stageNavFiles, err := filepath.Glob(fmt.Sprintf("static/stable/%s/navigation/quay-navigation.json", env))
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

			flatData, err := flattenLinks(navItemData)
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
	envIdex, err := flattenIndexBase(append(indexBase, staticBase...), env)

	if err != nil {
		return []ModuleIndexEntry{}, err
	}

	return envIdex, nil
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
		return bodyString, fmt.Errorf(bodyString)
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

	client := &http.Client{}

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
		return fmt.Errorf(bodyString)
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

	index, err := constructIndex(env)
	if err != nil {
		return err
	}

	err = uploadIndex(token, index, hydraHost)
	if err != nil {
		return err
	}

	return nil
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

	for _, env := range []SearchEnv{Stage, Prod} {
		fmt.Println("Attempt to publish search index for ", env, " environment.")
		err := deployIndex(env, secrets[env], ssoHosts[env], hydraHost[env])
		if err != nil {
			fmt.Println("Failed to deploy search index for ", env, "environment.")
			fmt.Println(err)
		}
	}

	fmt.Println("Search index published successfully")
}
