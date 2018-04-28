package db

import (
	"context"
	"database/sql"
	"time"
)

// source kinds.
const (
	SourceKindGit = "git"
)

// SourceTable lists repositories to pull kicad files from.
type SourceTable struct{}

// Setup is called on initialization to create necessary structures in the database.
func (t *SourceTable) Setup(ctx context.Context, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
  	CREATE TABLE IF NOT EXISTS sources (
  		rowid INTEGER PRIMARY KEY AUTOINCREMENT,
  	  kind STRING NOT NULL,
  	  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      url VARCHAR(512) NOT NULL
  	);
    CREATE UNIQUE INDEX IF NOT EXISTS sources_url ON sources(url);
	`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Source records a single source from which kc files are ingested.
type Source struct {
	UID       int       `json:"uid"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

// AddSource commits a new source record.
func AddSource(ctx context.Context, src *Source, db *sql.DB) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
    INSERT INTO
      sources (kind, url)
      VALUES (?, ?);`, src.Kind, src.URL)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetSources returns all sources.
func GetSources(ctx context.Context, db *sql.DB) ([]*Source, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, `
		SELECT rowid, kind, created_at, url FROM sources;
	`)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var output []*Source
	for res.Next() {
		var o Source
		if err := res.Scan(&o.UID, &o.Kind, &o.CreatedAt, &o.URL); err != nil {
			return nil, err
		}
		output = append(output, &o)
	}

	return output, nil
}
