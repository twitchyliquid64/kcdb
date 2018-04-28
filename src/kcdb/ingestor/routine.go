package ingestor

import (
	"fmt"
	"io/ioutil"
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

	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL: current.URL,
	})
	if err != nil {
		return err
	}

	return filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("[ingest][walk] Could not read %q: %v\n", path, err)
			return err
		}
		if strings.HasSuffix(path, ".kicad_mod") {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			footprint, err := mod.DecodeModule(strings.NewReader(string(b)))
			if err != nil {
				fmt.Printf("[ingest][footprint] Failed parsing %q: %v\n", path, err)
				return nil
			}
			fmt.Printf("[ingest][footprint] Successfully parsed %s\n", footprint.Name)
		}
		return nil
	})
}
