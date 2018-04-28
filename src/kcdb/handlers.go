package kcdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"kcdb/db"
	"kcdb/ingestor"
	"kcdb/mod"
	"kcdb/search"
)

// SearchHandler performs a search.
func SearchHandler(w http.ResponseWriter, req *http.Request) {
	var query struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		fmt.Printf("Err: %v\n", err)
		return
	}

	results, err := search.Search(req.Context(), query.Query)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	b, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// FootprintHandler serves the HTML for viewing a footprint
func FootprintHandler(w http.ResponseWriter, req *http.Request) {
	f, err := os.Open("static/part.html")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, f)
}

// ModuleDetails replies with a JSON blob representing the Module.
func ModuleDetails(w http.ResponseWriter, req *http.Request) {
	var raw []byte
	if strings.HasPrefix(req.URL.Path, "/module/details/") {
		fp, err := db.FootprintByURL(req.Context(), req.URL.Path[len("/module/details/"):], db.DB())
		if err != nil {
			if err == os.ErrNotExist {
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal error", http.StatusInternalServerError)
			}
			fmt.Printf("Err: %v\n", err)
			return
		}
		raw = fp.Data
	} else {
		http.Error(w, "The request did not indicate what footprint should be returned", http.StatusBadRequest)
		return
	}

	mod, err := mod.DecodeModule(strings.NewReader(string(raw)))
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}
	b, err := json.Marshal(mod)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// ListSources responds with a list of sources.
func ListSources(w http.ResponseWriter, req *http.Request) {
	sources, err := db.GetSources(req.Context(), db.DB())
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}
	b, err := json.Marshal(sources)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// IngestState responds with the current state of the ingestor.
func IngestState(w http.ResponseWriter, req *http.Request) {
	next, err := ingestor.ComputeIngestTargets()
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}
	current, delay, nextIngest := ingestor.State()
	b, err := json.Marshal(map[string]interface{}{
		"current":              current,
		"ingest_delay_seconds": delay,
		"next_ingest":          nextIngest,
		"next_sources":         next,
	})
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
