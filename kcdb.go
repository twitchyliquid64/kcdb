package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"kcdb"
	"kcdb/admin"
	"kcdb/db"
	"kcdb/ingestor"
)

var (
	listenerFlag    = flag.String("listener", "localhost:8080", "Address to listen on")
	updateDelayFlag = flag.Int("update-delay", 120, "Seconds between ingesting from sources")
	adminSecretFlag = flag.String("admin-secret", "", "Secret to use for admin RPCs")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	admin.SetSecret(*adminSecretFlag)

	initHandlers()
	_, err := db.Init(ctx, "kc.db")
	if err != nil {
		fmt.Printf("Failed to init db: %v\n", err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "add-git-source":
		newGitSource(ctx, flag.Arg(1))

	case "dump-sources":
		dumpSources(ctx)

	case "load-sources":
		loadSources(ctx)

	case "", "run":
		if err := ingestor.Start(*updateDelayFlag); err != nil {
			fmt.Printf("Failed to setup ingestor: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Now listening on %s\n", *listenerFlag)
		makeServer().ListenAndServe()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %q\n", flag.Arg(0))
		os.Exit(1)
	}
}

func dumpSources(ctx context.Context) {
	s, err := db.GetSources(ctx, db.DB())
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetSources() failed: %v\n", err)
		os.Exit(1)
	}
	j, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println(string(j))
}

func loadSources(ctx context.Context) {
	d, err := ioutil.ReadFile(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadFile() failed: %v\n", err)
		os.Exit(1)
	}
	var sources []db.Source
	if err := json.Unmarshal(d, &sources); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode sources: %v\n", err)
		os.Exit(1)
	}
	for _, source := range sources {
		fmt.Printf("[%.3d] %s ... ", source.UID, source.URL)

		s, err := db.GetSource(ctx, source.UID, db.DB())
		if err == os.ErrNotExist || s.URL != source.URL {
			if err := db.CreateSource(ctx, &source, db.DB()); err != nil {
				fmt.Println("Err!\n\t" + err.Error())
			} else {
				fmt.Println("Created!")
			}
			continue
		}

		fmt.Println("Exists.")
	}
}

func newGitSource(ctx context.Context, url string) {
	err := db.AddSource(ctx, &db.Source{
		Kind: db.SourceKindGit,
		URL:  url,
	}, db.DB())
	if err != nil {
		fmt.Printf("Failed to add git source: %v\n", err)
		os.Exit(1)
	}
}

func makeServer() *http.Server {
	// make our server objects
	s := &http.Server{
		Addr: *listenerFlag,
	}
	return s
}

func initHandlers() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/module/details", kcdb.ModuleDetails)
	http.HandleFunc("/module/details/", kcdb.ModuleDetails)
	http.HandleFunc("/footprint/", kcdb.FootprintHandler)
	http.HandleFunc("/symbol/", kcdb.SymbolHandler)
	http.HandleFunc("/sym/details/", kcdb.SymbolDetails)
	http.HandleFunc("/sources/all", kcdb.ListSources)
	http.HandleFunc("/search/all", kcdb.SearchHandler)
	http.HandleFunc("/ingestor/status", kcdb.IngestState)
	http.HandleFunc("/admin/sources/params", admin.UpdateSourceAdmin)
	http.HandleFunc("/admin/sources/add", admin.AddSourceAdmin)
}
