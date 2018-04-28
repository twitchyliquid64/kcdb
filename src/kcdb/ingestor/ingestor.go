package ingestor

import (
	"context"
	"kcdb/db"
	"sync"
	"time"
)

var lock sync.Mutex
var nextIngest time.Time
var ingestDelaySeconds int
var current *db.Source

// Start begins the ingestion routine.
func Start(delaySecs int) error {
	ingestDelaySeconds = delaySecs
	nextIngest = time.Now().Add(time.Duration(ingestDelaySeconds) * time.Second / 2)

	if err := clearDir(); err != nil {
		return err
	}

	go ingestRoutine()
	return nil
}

// ComputeIngestTargets returns the source which should next be ingested.
func ComputeIngestTargets() ([]*db.Source, error) {
	lock.Lock()
	defer lock.Unlock()

	sources, err := db.SourcesLastUpdated(context.Background(), 5, db.DB())
	if err != nil {
		return nil, err
	}
	if len(sources) > 1 && current != nil && sources[0].UID == current.UID {
		sources = sources[1:]
	}
	return sources, nil
}

// State returns the internal state of the ingestor.
func State() (*db.Source, int, time.Time) {
	lock.Lock()
	defer lock.Unlock()
	return current, ingestDelaySeconds, nextIngest
}

func ingestRoutine() {
	for {
		time.Sleep(time.Second)
		if nextIngest.Before(time.Now()) {
			doIngest()
			lock.Lock()
			nextIngest = time.Now().Add(time.Duration(ingestDelaySeconds) * time.Second)
			current = nil
			lock.Unlock()
		}
	}
}
