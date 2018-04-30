package db

import (
	"context"
	"database/sql"
	"os"
	"strings"
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
			name VARCHAR(128) NOT NULL,
			pin_count INT NOT NULL,
			attr VARCHAR(32) NOT NULL,
			tags VARCHAR(256) NOT NULL,
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

	Name     string `json:"name"`
	PinCount int    `json:"pin_count"`
	Attr     string `json:"attr"`
	Tags     string `json:"tags"`

	// Not stored in DB
	Rank int `json:"rank,omitempty"`
}

// MakeFootprintURL creates a pretty URL for the footprint.
func MakeFootprintURL(repoURL, repoPath string) string {
	if repoURL[len(repoURL)-1] == '/' {
		repoURL = repoURL[:len(repoURL)-2]
	}
	repoURL = strings.TrimPrefix(repoURL, "https://")
	repoURL = strings.TrimPrefix(repoURL, "http://")
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
    UPDATE footprints SET data=?, pin_count=?, name=?, attr=?, tags=?, updated_at=CURRENT_TIMESTAMP WHERE rowid = ?;`, fp.Data, fp.PinCount, fp.Name, fp.Attr, fp.Tags, fp.UID)
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
      footprints (source_id, url, data, name, pin_count, attr, tags)
      VALUES (?, ?, ?, ?, ?, ?, ?);`, fp.SourceID, fp.URL, fp.Data, fp.Name, fp.PinCount, fp.Attr, fp.Tags)
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

// FootprintByURL returns the specified footprint
func FootprintByURL(ctx context.Context, url string, db *sql.DB) (*Footprint, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, `
    SELECT rowid, source_id, updated_at, url, data, name, pin_count, attr, tags FROM footprints WHERE url = ?;
  `, url)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if !res.Next() {
		return nil, os.ErrNotExist
	}
	var fp Footprint
	return &fp, res.Scan(&fp.UID, &fp.SourceID, &fp.UpdatedAt, &fp.URL, &fp.Data, &fp.Name, &fp.PinCount, &fp.Attr, &fp.Tags)
}

// FpSearchParam specifies parameters to constrain a footprint search.
type FpSearchParam struct {
	Keywords []string
	PinCount int
	Attr     string
}

// FootprintSearch performs a footprint search
func FootprintSearch(ctx context.Context, search FpSearchParam, db *sql.DB) ([]*Footprint, error) {
	where := ""
	params := []interface{}{}
	for i, kw := range search.Keywords {
		where += "(name LIKE ? OR tags LIKE ?)"
		params = append(params, "%"+kw+"%", "%"+kw+"%")
		if i < (len(search.Keywords) - 1) {
			where += " AND "
		}
	}
	if search.Attr != "" {
		where += " AND attr LIKE ?"
		params = append(params, search.Attr)
	}
	if search.PinCount != 0 {
		where += " AND pin_count = ?"
		params = append(params, search.PinCount)
	}

	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, "SELECT rowid, source_id, updated_at, url, name, pin_count, attr, tags FROM footprints WHERE "+where+" LIMIT 50;", params...)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var out []*Footprint
	for res.Next() {
		var fp Footprint
		if err := res.Scan(&fp.UID, &fp.SourceID, &fp.UpdatedAt, &fp.URL, &fp.Name, &fp.PinCount, &fp.Attr, &fp.Tags); err != nil {
			return nil, err
		}
		out = append(out, &fp)
	}

	return out, nil
}
