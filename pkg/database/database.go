package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

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

func (this *Database) RefreshBusStopTable(busStops []BusStopTable) {
	db, err := sql.Open(dbType, this.dbPath)
	errCheck("[Database][ResetBusStopTable] Cannot open DB", err)
	defer db.Close()

	queryDropTable := "DROP TABLE IF EXISTS bus_stop_info"
	queryCreateTable := "CREATE TABLE IF NOT EXISTS bus_stop_info(BusStopCode TEXT PRIMARY KEY, RoadName TEXT, Description TEXT, Latitude float(24), Longitude float(24))"

	transaction, err := db.Begin()
	errCheck("[Database][ResetBusStopTable] Problems preparing transaction", err)

	defer func() {
		if err != nil {
			err = transaction.Rollback()
			errCheck("[Database][ResetBusStopTable] Problems with rollback", err)
		}
	}()

	_, err = transaction.Exec(queryDropTable)
	errCheck("[Database][ResetBusStopTable] Error with dropping table", err)

	_, err = transaction.Exec(queryCreateTable)
	errCheck("[Database][ResetBusStopTable] Error with create table", err)

	batchCounter := 0
	batchLimit := 100

	queryInsertIntoStr := make([]string, 0, batchLimit)
	queryInsertIntoArgs := make([]interface{}, 0, batchLimit*5)

	for _, busStop := range busStops {
		queryInsertIntoStr = append(queryInsertIntoStr, "(?,?,?,?,?)")
		queryInsertIntoArgs = append(queryInsertIntoArgs, busStop.BusStopCode)
		queryInsertIntoArgs = append(queryInsertIntoArgs, busStop.RoadName)
		queryInsertIntoArgs = append(queryInsertIntoArgs, busStop.Description)
		queryInsertIntoArgs = append(queryInsertIntoArgs, busStop.Latitude)
		queryInsertIntoArgs = append(queryInsertIntoArgs, busStop.Longitude)
		batchCounter++

		if batchCounter == batchLimit {
			queryInsertInto := fmt.Sprintf("INSERT INTO bus_stop_info (BusStopCode, RoadName, Description, Latitude, Longitude) VALUES %s", strings.Join(queryInsertIntoStr, ","))
			_, err = transaction.Exec(queryInsertInto, queryInsertIntoArgs...)
			errCheck("[Database][ResetBusStopTable] Error with insert into", err)
			queryInsertIntoStr = queryInsertIntoStr[:0]
			queryInsertIntoArgs = queryInsertIntoArgs[:0]
			batchCounter = 0
		}
	}

	// Commit
	err = transaction.Commit()
	errCheck("[Database][ResetBusStopTable] Error committing transaction", err)
}

func (this *Database) GetBusStop(busStop string) *BusStopTable {
	db, err := sql.Open(dbType, this.dbPath)
	errCheck("[Database][ResetBusStopTable] Cannot open DB", err)
	defer db.Close()

	query := "SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info WHERE BusStopCode = ?"
	rows, err := db.Query(query, busStop)
	errCheck("[Database][DoesBusStopExist] Problem with query", err)

	// I only expect 1 row
	if !rows.Next() {
		return nil
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret
}
