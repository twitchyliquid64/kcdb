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
			updated_at TIMESTAMP NOT NULL DEFAULT 0,
      url VARCHAR(512) NOT NULL,
			ranking_priority INT NOT NULL DEFAULT 0,
			tag VARCHAR(128) NOT NULL DEFAULT '',
			metadata VARCHAR(2048) NOT NULL DEFAULT '{}'
  	);
    CREATE UNIQUE INDEX IF NOT EXISTS sources_url ON sources(url);
	`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return t.migratev1(ctx, db)
}

func (t *SourceTable) migratev1(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "SELECT tag FROM sources LIMIT 1;")
	if err == nil {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE sources
		ADD COLUMN ranking_priority INT NOT NULL DEFAULT 0;`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE sources
		ADD COLUMN tag VARCHAR(128) NOT NULL DEFAULT '';`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE sources
		ADD COLUMN metadata VARCHAR(2048) NOT NULL DEFAULT '{}';`)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// Source records a single source from which kc files are ingested.
type Source struct {
	UID       int       `json:"uid"`
	Kind      string    `json:"kind"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	URL       string    `json:"url"`
	Rank 			int				`json:"rank"`
	Tag       string    `json:"tag"`
	Metadata  string    `json:"metadata"`
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

// SourcesLastUpdated returns sources in order of least-recently updated.
func SourcesLastUpdated(ctx context.Context, limit int, db *sql.DB) ([]*Source, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, `
		SELECT rowid, kind, created_at, updated_at, url, ranking_priority, tag, metadata FROM sources ORDER BY updated_at ASC LIMIT ?;
	`, limit)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var output []*Source
	for res.Next() {
		var o Source
		if err := res.Scan(&o.UID, &o.Kind, &o.CreatedAt, &o.UpdatedAt, &o.URL, &o.Rank, &o.Tag, &o.Metadata); err != nil {
			return nil, err
		}
		output = append(output, &o)
	}

	return output, nil
}

// SetSourceUpdated updates the updated_at value of the given source to the current time.
func SetSourceUpdated(ctx context.Context, uid int, db *sql.DB) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE sources SET updated_at=CURRENT_TIMESTAMP WHERE rowid = ?;`, uid)
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
		SELECT rowid, kind, created_at, updated_at, url, ranking_priority, tag, metadata FROM sources;
	`)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var output []*Source
	for res.Next() {
		var o Source
		if err := res.Scan(&o.UID, &o.Kind, &o.CreatedAt, &o.UpdatedAt, &o.URL, &o.Rank, &o.Tag, &o.Metadata); err != nil {
			return nil, err
		}
		output = append(output, &o)
	}

	return output, nil
}
