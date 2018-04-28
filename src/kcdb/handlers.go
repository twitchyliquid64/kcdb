package kcdb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"kcdb/db"
	"kcdb/mod"
	"kcdb/ingestor"
)

// ModuleDetails replies with a JSON blob representing the Module.
func ModuleDetails(w http.ResponseWriter, req *http.Request) {
	f, err := ioutil.ReadFile("static/testdata/SOIC-20_W7.5mm.kicad_mod")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	mod, err := mod.DecodeModule(strings.NewReader(string(f)))
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

func IngestState(w http.ResponseWriter, req *http.Request) {
	current, delay, nextIngest := ingestor.State()
	b, err := json.Marshal(map[string]interface{}{
		"current": current,
		"ingest_delay_seconds": delay,
		"next_source": nextIngest,
	})
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		fmt.Printf("Err: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
