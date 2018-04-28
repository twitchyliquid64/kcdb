package db

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/mattn/go-sqlite3"
)

var dbGlobal *sql.DB

// DatabaseTable represents the manager object for a database table.
type DatabaseTable interface {
	Setup(ctx context.Context, db *sql.DB) error
}

var tables = []DatabaseTable{
	&SourceTable{},
}

// Init is called with database information to initialise a database session, creating any necessary tables.
func Init(ctx context.Context, connString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}
	dbGlobal = db

	for _, table := range tables {
		err := table.Setup(ctx, db)
		if err != nil {
			fmt.Printf("Problem initialising %q: %v\n", reflect.TypeOf(table), err)
			db.Close()
			return nil, err
		}
	}

	return db, nil
}

// DB returns the database handle.
func DB() *sql.DB {
	if dbGlobal == nil {
		panic("db not initialized")
	}
	return dbGlobal
}
