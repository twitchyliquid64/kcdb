package db

import (
	"context"
	"database/sql"
	"fmt"
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
      data BLOB NOT NULL,
			pin_count INT NOT NULL DEFAULT 0,
			condensed_pins VARCHAR(32768) NOT NULL DEFAULT ''
  	);
		CREATE UNIQUE INDEX IF NOT EXISTS symbols_url ON symbols(url);
	`)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return t.migratev1(ctx, db)
}

func (t *SymbolTable) migratev1(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "SELECT pin_count FROM symbols LIMIT 1;")
	if err == nil {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE symbols
		ADD COLUMN pin_count INT NOT NULL DEFAULT 0;`)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`ALTER TABLE symbols
		ADD COLUMN condensed_pins VARCHAR(32768) NOT NULL DEFAULT '';`)
	if err != nil {
		return err
	}
	return tx.Commit()
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
	PinData   string `json:"pin_data"`

	PinCount int `json:"pin_count"`
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
    UPDATE symbols SET data=?, name=?, condensed_fields=?, pin_count=?, condensed_pins=?, updated_at=CURRENT_TIMESTAMP WHERE rowid = ?;`, sym.Data, sym.Name, sym.FieldData, sym.PinCount, sym.PinData, sym.UID)
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
      symbols (source_id, url, data, name, condensed_fields, pin_count, condensed_pins)
      VALUES (?, ?, ?, ?, ?, ?, ?);`, sym.SourceID, sym.URL, sym.Data, sym.Name, sym.FieldData, sym.PinCount, sym.PinData)
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

// SymSearchParam specifies parameters to constrain a symbol search.
type SymSearchParam struct {
	Keywords []string
	PinCount int
}

// SymbolSearch performs a symbol search.
func SymbolSearch(ctx context.Context, search SymSearchParam, db *sql.DB) ([]*Symbol, error) {
	where := ""
	params := []interface{}{}
	for i, kw := range search.Keywords {
		where += "(name LIKE ? OR condensed_fields LIKE ? OR condensed_pins LIKE ?)"
		params = append(params, "%"+kw+"%", "%"+kw+"%", "%"+kw+"%")
		if i < (len(search.Keywords) - 1) {
			where += " AND "
		}
	}
	if search.PinCount != 0 {
		where += " AND pin_count = ?"
		params = append(params, search.PinCount)
	}

	dbLock.RLock()
	defer dbLock.RUnlock()

	res, err := db.QueryContext(ctx, "SELECT rowid, source_id, updated_at, url, name, pin_count FROM symbols WHERE "+where+" LIMIT 65;", params...)
	if err != nil {
		fmt.Printf("db.QueryContext(%q) failed: %v\n", "... WHERE "+where, err)
		return nil, err
	}
	defer res.Close()

	var out []*Symbol
	for res.Next() {
		var sym Symbol
		if err := res.Scan(&sym.UID, &sym.SourceID, &sym.UpdatedAt, &sym.URL, &sym.Name, &sym.PinCount); err != nil {
			fmt.Printf("db.Scan(%q) failed: %v\n", "... WHERE "+where, err)
			return nil, err
		}
		out = append(out, &sym)
	}

	return out, nil
}
