package ingestor

import (
	"kcdb/db"
	"sync"
	"time"
)

var lock sync.Mutex
var nextIngest time.Time
var ingestDelaySeconds int
var current *db.Source

// Start begins the ingestion routine.
func Start(delaySecs int) {
	ingestDelaySeconds = delaySecs
}

// ComputeIngestTarget returns the source which should next be ingested.
func ComputeIngestTarget() {
	lock.Lock()
	defer lock.Unlock()
}

// State returns the internal state of the ingestor.
func State() (*db.Source, int, time.Time) {
	lock.Lock()
	defer lock.Unlock()
	return current, ingestDelaySeconds, nextIngest
}
