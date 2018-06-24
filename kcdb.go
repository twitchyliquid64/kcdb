package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	//"io/ioutil"

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

	if flag.Arg(0) == "add-git-source" {
		newGitSource(ctx, flag.Arg(1))
		return
	}

	if err := ingestor.Start(*updateDelayFlag); err != nil {
		fmt.Printf("Failed to setup ingestor: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Now listening on %s\n", *listenerFlag)
	makeServer().ListenAndServe()
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
	http.HandleFunc("/sources/all", kcdb.ListSources)
	http.HandleFunc("/search/all", kcdb.SearchHandler)
	http.HandleFunc("/ingestor/status", kcdb.IngestState)
	http.HandleFunc("/admin/sources/params", admin.UpdateSourceAdmin)
	http.HandleFunc("/admin/sources/add", admin.AddSourceAdmin)
}
