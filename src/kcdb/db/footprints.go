package db

import (
	"context"
	"database/sql"
	"time"
)

// FootprintTable contains footprints.
type FootprintTable struct{}

// Setup is called on initialization to create necessary structures in the database.
func (t *FootprintTable) Setup(ctx context.Context, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
  	CREATE TABLE IF NOT EXISTS footprints (
  		rowid INTEGER PRIMARY KEY AUTOINCREMENT,
  	  source_id INT NOT NULL,
      url VARCHAR(1024) NOT NULL,
  	  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
      data BLOB NOT NULL
  	);
    CREATE UNIQUE INDEX IF NOT EXISTS footprints_url ON footprints(url);
	`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Footprint contains information about a footprint.
type Footprint struct {
	UID       int       `json:"uid"`
	SourceID  int       `json:"source_uid"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      []byte    `json:"data,string"`
}

// MakeFootprintURL creates a pretty URL for the footprint.
func MakeFootprintURL(repoURL, repoPath string) string {
	if repoURL[len(repoURL)-1] == '/' {
		repoURL = repoURL[:len(repoURL)-2]
	}
	return repoURL + "::" + repoPath
}

// FootprintExists identifies if a footprint is stored with that URL.
func FootprintExists(ctx context.Context, url string, db *sql.DB) (bool, int, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, `
    SELECT rowid FROM footprints WHERE url = ?;
  `, url)
	if err != nil {
		return false, 0, err
	}
	defer res.Close()
	if !res.Next() {
		return false, 0, nil
	}
	var id int
	if err := res.Scan(&id); err != nil {
		return false, 0, err
	}
	return true, id, nil
}

// UpdateFootprint updates a footprint.
func UpdateFootprint(ctx context.Context, fp *Footprint, db *sql.DB) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
    UPDATE footprints SET data=?, updated_at=CURRENT_TIMESTAMP WHERE rowid = ?;`, fp.Data, fp.UID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// CreateFootprint creates a footprint.
func CreateFootprint(ctx context.Context, fp *Footprint, db *sql.DB) (int, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	e, err := tx.ExecContext(ctx, `
    INSERT INTO
      footprints (source_id, url, data)
      VALUES (?, ?, ?);`, fp.SourceID, fp.URL, fp.Data)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	id, err := e.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
