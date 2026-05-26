package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"slices"

	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

const specsFilePath = "static/specs-generated.json"

type apiDocEntry struct {
	URL          string          `json:"url"`
	BundleLabels []string        `json:"bundleLabels"`
	Spec         json.RawMessage `json:"spec"`
}

func GetApiDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=3600")

	bundleFilter := r.URL.Query().Get("bundle")

	if bundleFilter == "" {
		serveSpecsFile(w)
		return
	}

	serveFilteredSpecs(w, bundleFilter)
}

func serveSpecsFile(w http.ResponseWriter) {
	f, err := os.Open(specsFilePath)
	if err != nil {
		handleFileError(w, err)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		logrus.Errorf("failed to stream specs file %s: %v", specsFilePath, err)
	}
}

func serveFilteredSpecs(w http.ResponseWriter, bundle string) {
	data, err := os.ReadFile(specsFilePath)
	if err != nil {
		handleFileError(w, err)
		return
	}

	var allSpecs map[string][]apiDocEntry
	if err := json.Unmarshal(data, &allSpecs); err != nil {
		logrus.Errorf("failed to parse specs file %s: %v", specsFilePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(util.ErrorResponse{Errors: []string{"failed to parse specs file"}})
		return
	}

	filtered := make(map[string][]apiDocEntry)
	for name, entries := range allSpecs {
		for _, entry := range entries {
			if slices.Contains(entry.BundleLabels, bundle) {
				filtered[name] = append(filtered[name], entry)
			}
		}
	}

	json.NewEncoder(w).Encode(filtered)
}

func handleFileError(w http.ResponseWriter, err error) {
	if errors.Is(err, os.ErrNotExist) {
		logrus.Warnf("specs file not found at %s, returning empty object", specsFilePath)
		if _, writeErr := w.Write([]byte("{}")); writeErr != nil {
			logrus.Errorf("failed to write fallback response: %v", writeErr)
		}
		return
	}
	logrus.Errorf("failed to open specs file %s: %v", specsFilePath, err)
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(util.ErrorResponse{Errors: []string{"failed to read specs file"}})
}

func MakeApiDocsRoutes(sub chi.Router) {
	sub.Get("/", GetApiDocs)
}
