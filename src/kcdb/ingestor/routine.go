package ingestor

import (
	"context"
	"fmt"
	"io/ioutil"
	"kcdb/db"
	"kcdb/mod"
	"os"
	"path/filepath"
	"strings"

	git "gopkg.in/src-d/go-git.v4"
)

const tmpDir = "/tmp/kcdb_repo"

func clearDir() error {
	if err := os.RemoveAll(tmpDir); err != nil && os.IsNotExist(err) {
		return err
	}
	return os.Mkdir(tmpDir, 0755)
}

func doIngest() error {
	targets, err := ComputeIngestTargets()
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}
	lock.Lock()
	current = targets[0]
	lock.Unlock()

	if err = clearDir(); err != nil {
		return err
	}
	defer func() {
		fmt.Printf("[ingest] Starting Vacuum.\n")
		db.Vacuum(db.DB())
		fmt.Printf("[ingest] Finished routine.\n")
	}()

	fmt.Printf("[ingest][clone] Cloning: %v (%d)\n", current.URL, current.UID)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: current.URL,
	})
	if err != nil {
		return err
	}
	fmt.Printf("[ingest][clone] Finished.\n")

	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[ingest][walk] Could not read %q: %v\n", path, err)
			return err
		}
		if strings.HasSuffix(path, ".kicad_mod") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			url := db.MakeFootprintURL(current.URL, path[len(tmpDir)+1:])

			footprint, err := mod.DecodeModule(strings.NewReader(string(b)))
			if err != nil {
				fmt.Printf("[ingest][footprint] Failed parsing %q: %v\n", path, err)
				fmt.Println(string(b))
				return nil
			}

			_, err = upsertFootprint(current, url, b, footprint)
			if err != nil {
				return err
			}

			//fmt.Printf("[ingest][footprint] Successfully parsed %s\n", footprint.Name)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return db.SetSourceUpdated(context.Background(), current.UID, db.DB())
}

func upsertFootprint(source *db.Source, url string, b []byte, fp *mod.Module) (int, error) {
	ctx := context.Background()
	exists, uid, err := db.FootprintExists(ctx, url, db.DB())
	if err != nil {
		return 0, err
	}
	if exists {
		return uid, db.UpdateFootprint(ctx, &db.Footprint{UID: uid,
			Data:     b,
			URL:      url,
			SourceID: source.UID,
			PinCount: len(fp.Pads),
			Name:     fp.Name,
			Attr:     strings.Join(fp.Attrs, ","),
		}, db.DB())
	}
	return db.CreateFootprint(ctx, &db.Footprint{
		Data:     b,
		URL:      url,
		SourceID: source.UID,
		PinCount: len(fp.Pads),
		Name:     fp.Name,
		Attr:     strings.Join(fp.Attrs, ","),
	}, db.DB())
}
