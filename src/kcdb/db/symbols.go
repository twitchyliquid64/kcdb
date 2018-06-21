package db

import (
	"context"
	"database/sql"
	"time"
)

// SymbolTable contains symbols.
type SymbolTable struct{}

// Setup is called on initialization to create necessary structures in the database.
func (t *SymbolTable) Setup(ctx context.Context, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
  	CREATE TABLE IF NOT EXISTS symbols (
  		rowid INTEGER PRIMARY KEY AUTOINCREMENT,
  	  source_id INT NOT NULL,
      url VARCHAR(1024) NOT NULL,
  	  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			name VARCHAR(128) NOT NULL,
			condensed_fields VARCHAR(32768) NOT NULL,
      data BLOB NOT NULL
  	);
		CREATE UNIQUE INDEX IF NOT EXISTS symbols_url ON symbols(url);
	`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// Symbol contains information about a symbol.
type Symbol struct {
	UID       int       `json:"uid"`
	SourceID  int       `json:"source_uid"`
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
	Data      []byte    `json:"data,string"`

	Name      string `json:"name"`
	FieldData string `json:"field_data"`
	// Not stored in DB
	Rank int `json:"rank,omitempty"`
}

// SymbolExists identifies if a symbol is stored with that URL.
func SymbolExists(ctx context.Context, url string, db *sql.DB) (bool, int, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, `
    SELECT rowid FROM symbols WHERE url = ?;
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

// UpdateSymbol updates a symbol.
func UpdateSymbol(ctx context.Context, sym *Symbol, db *sql.DB) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
    UPDATE symbols SET data=?, name=?, condensed_fields=?, updated_at=CURRENT_TIMESTAMP WHERE rowid = ?;`, sym.Data, sym.Name, sym.FieldData, sym.UID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// CreateSymbol creates a symbol.
func CreateSymbol(ctx context.Context, sym *Symbol, db *sql.DB) (int, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	e, err := tx.ExecContext(ctx, `
    INSERT INTO
      symbols (source_id, url, data, name, condensed_fields)
      VALUES (?, ?, ?, ?, ?);`, sym.SourceID, sym.URL, sym.Data, sym.Name, sym.FieldData)
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
