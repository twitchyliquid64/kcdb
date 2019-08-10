package ingestor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"kcdb/db"
	"kcdb/sym"
	"os"
	"path/filepath"
	"strings"

	"github.com/twitchyliquid64/kcgen/pcb"
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
	defer db.SetSourceUpdated(context.Background(), current.UID, db.DB())
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
			url := db.MakePartURL(current.URL, path[len(tmpDir)+1:])

			//fmt.Printf("File: %+v\n", path)
			mod, err := pcb.ParseModule(strings.NewReader(string(b)))
			if err != nil {
				fmt.Printf("[ingest][footprint] Failed parsing %q: %v\n", path, err)
				fmt.Println(string(b))
				return nil
			}

			_, err = upsertFootprint(current, url, b, mod)
			if err != nil {
				return err
			}

			//fmt.Printf("[ingest][footprint] Successfully parsed %s\n", footprint.Name)
		} else if strings.HasSuffix(path, ".lib") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			url := db.MakePartURL(current.URL, path[len(tmpDir)+1:])

			symbols, err := sym.DecodeSymbolLibrary(bytes.NewBuffer(b))
			if err != nil {
				fmt.Printf("[ingest][symbols] Failed parsing %q: %v\n", path, err)
				// fmt.Println(string(b))
				return nil
			}

			for i := range symbols {
				_, err = upsertSymbol(current, url+"::"+symbols[i].Name, []byte(symbols[i].RawData), symbols[i])
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func upsertFootprint(source *db.Source, url string, b []byte, fp *pcb.Module) (int, error) {
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
			Tags:     strings.Join(fp.Tags, ","),
		}, db.DB())
	}
	return db.CreateFootprint(ctx, &db.Footprint{
		Data:     b,
		URL:      url,
		SourceID: source.UID,
		PinCount: len(fp.Pads),
		Name:     fp.Name,
		Attr:     strings.Join(fp.Attrs, ","),
		Tags:     strings.Join(fp.Tags, ","),
	}, db.DB())
}

func upsertSymbol(source *db.Source, url string, b []byte, s *sym.Symbol) (int, error) {
	ctx := context.Background()
	exists, uid, err := db.SymbolExists(ctx, url, db.DB())
	if err != nil {
		return 0, err
	}

	fieldData := ""
	for i := range s.Fields {
		if s.Fields[i].Value == "" {
			continue
		}
		fieldData += s.Fields[i].Value
		if i < (len(s.Fields) - 1) {
			fieldData += " "
		}
	}

	pinData := ""
	for i := range s.Pins {
		if s.Pins[i].Name == "" {
			continue
		}
		pinData += s.Pins[i].Name
		if i < (len(s.Pins) - 1) {
			pinData += " "
		}
	}

	if exists {
		return uid, db.UpdateSymbol(ctx, &db.Symbol{
			UID:       uid,
			Data:      b,
			URL:       url,
			SourceID:  source.UID,
			Name:      s.Name,
			FieldData: fieldData,
			PinCount:  len(s.Pins),
			PinData:   pinData,
		}, db.DB())
	}
	return db.CreateSymbol(ctx, &db.Symbol{
		UID:       uid,
		Data:      b,
		URL:       url,
		SourceID:  source.UID,
		Name:      s.Name,
		FieldData: fieldData,
		PinCount:  len(s.Pins),
		PinData:   pinData,
	}, db.DB())
}
