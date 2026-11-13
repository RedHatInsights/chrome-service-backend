package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"

	requestlogger "github.com/RedHatInsights/chrome-service-backend/rest/logger"
	"github.com/RedHatInsights/chrome-service-backend/rest/util"
	"github.com/go-chi/chi/v5"
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
		serveSpecsFile(w, r)
		return
	}

	serveFilteredSpecs(w, r, bundleFilter)
}

func serveSpecsFile(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(specsFilePath)
	if err != nil {
		handleFileError(w, r, err)
		return
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		requestlogger.LogFor(r.Context()).WithField("error_type", fmt.Sprintf("%T", err)).Error("specs response streaming failed")
	}
}

func serveFilteredSpecs(w http.ResponseWriter, r *http.Request, bundle string) {
	data, err := os.ReadFile(specsFilePath)
	if err != nil {
		handleFileError(w, r, err)
		return
	}

	var allSpecs map[string][]apiDocEntry
	if err := json.Unmarshal(data, &allSpecs); err != nil {
		requestlogger.LogFor(r.Context()).WithField("error_type", fmt.Sprintf("%T", err)).Error("specs file parsing failed")
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

	if err := json.NewEncoder(w).Encode(filtered); err != nil {
		requestlogger.LogFor(r.Context()).WithField("error_type", fmt.Sprintf("%T", err)).Error("filtered specs response encoding failed")
	}
}

func handleFileError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, os.ErrNotExist) {
		if _, writeErr := w.Write([]byte("{}")); writeErr != nil {
			requestlogger.LogFor(r.Context()).WithField("error_type", fmt.Sprintf("%T", writeErr)).Error("empty specs fallback response failed")
		}
		return
	}
	requestlogger.LogFor(r.Context()).WithField("error_type", fmt.Sprintf("%T", err)).Error("specs file read failed")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(util.ErrorResponse{Errors: []string{"failed to read specs file"}})
}

func MakeApiDocsRoutes(sub chi.Router) {
	sub.Get("/", GetApiDocs)
}
