package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	//"io/ioutil"

	"kcdb"
	"kcdb/db"
)

var (
	listenerFlag = flag.String("listener", "localhost:8080", "Address to listen on")
)

func main() {
	flag.Parse()
	ctx := context.Background()

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
	http.HandleFunc("/sources/all", kcdb.ListSources)
}
