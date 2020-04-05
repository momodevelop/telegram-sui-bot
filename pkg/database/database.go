package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Note: In this design, we don't close DB until application is closed
// This was suggested in sqlite3 documentation...apparently...

func errCheck(msg string, err error) {
	if err != nil {
		log.Printf("%s", msg)
		log.Panic(err)
	}
}

type Database struct {
	db *sql.DB
}

func New(dbPath string) *Database {
	log.Println("Checking database...")
	db, err := sql.Open("sqlite3", dbPath)
	errCheck("Something went wrong", err)

	ret := Database{
		db: db,
	}

	return &ret
}

func (this *Database) Close() {
	this.db.Close()
}
