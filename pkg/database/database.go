package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var dbType string = "sqlite3"

func errCheck(msg string, err error) {
	if err != nil {
		log.Printf("%s", msg)
		log.Panic(err)
	}
}

type Database struct {
	dbPath string
}

func New(dbPath string) *Database {
	log.Println("Checking database...")
	db, err := sql.Open(dbType, dbPath)
	errCheck("[New] Cannot open DB", err)
	defer db.Close()

	ret := Database{
		dbPath: dbPath,
	}

	return &ret
}

func (this *Database) ResetBusStopTable() {
	db, err := sql.Open(dbType, this.dbPath)
	errCheck("[Database][ResetBusStopTable] Cannot open DB", err)
	defer db.Close()

	queries := make([]string, 0)

	queries = append(queries, "DROP TABLE IF EXISTS bus_stop_info")
	queries = append(queries, "CREATE TABLE IF NOT EXISTS bus_stop_info(Id INT PRIMARY KEY, BusStopCode TEXT, RoadName TEXT, Description TEXT, Latitude float(24), Longitude float(24))")
	transaction, err := db.Begin()
	errCheck("[Database][ResetBusStopTable] Problems preparing transaction", err)

	defer func() {
		if err != nil {
			err = transaction.Rollback()
			errCheck("[Database][ResetBusStopTable] Problems with rollback", err)
		}
	}()

	for _, query := range queries {
		_, err := transaction.Exec(query)
		log.Printf("[Database][ResetBusStopTable] Executing: %s ", query)
		errCheck("[Database][ResetBusStopTable] Error with query: "+query, err)
	}

	err = transaction.Commit()
	errCheck("[Database][ResetBusStopTable] Error committing transaction", err)
}

func (this *Database) InsertBusStopTables() {

	/*query := "INSERT INTO bus_stop_info (Id, BusStopCode, RoadName, Description, Latitude, Longitude)"

	statement, err := this.db.Prepare(" CREATE_TABLE bus_stop_info")
	errCheck("Cannot drop bus_stop_info", err)
	_, err = statement.Exec()
	errCheck("Cannot execute DROP bus_stop_info", err)*/
}
