package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var dbType string = "sqlite3"

type Database struct {
	dbPath string
}

func New(dbPath string) *Database {
	log.Println("Checking database...")
	db, err := sql.Open(dbType, dbPath)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	ret := Database{
		dbPath: dbPath,
	}

	return &ret
}

func (this *Database) RefreshBusStopTable(busStops []BusStopTable) {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	queryDropTable := "DROP TABLE IF EXISTS bus_stop_info"
	queryCreateTable := "CREATE TABLE IF NOT EXISTS bus_stop_info(BusStopCode TEXT PRIMARY KEY, RoadName TEXT, Description TEXT, Latitude float(24), Longitude float(24))"

	transaction, err := db.Begin()
	if err != nil {
		log.Panic(err)
	}

	defer func() {
		if err != nil {
			err = transaction.Rollback()
			if err != nil {
				log.Panic(err)
			}
		}
	}()

	_, err = transaction.Exec(queryDropTable)
	if err != nil {
		log.Panic(err)
	}

	_, err = transaction.Exec(queryCreateTable)
	if err != nil {
		log.Panic(err)
	}

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
			if err != nil {
				log.Panic(err)
			}
			queryInsertIntoStr = queryInsertIntoStr[:0]
			queryInsertIntoArgs = queryInsertIntoArgs[:0]
			batchCounter = 0
		}
	}

	// Commit
	err = transaction.Commit()
	if err != nil {
		log.Panic(err)
	}
}

func (this *Database) GetBusStopByNearestLocation(latitude float64, longitude float64) *BusStopTable {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info ORDER BY (Latitude - ?) * (Latitude - ?) + (Longitude - ?) * (Longitude - ?)")
	rows, err := db.Query(query, latitude, latitude, longitude, longitude)
	if err != nil {
		log.Panic(err)
	}

	// I only expect 1 row
	if !rows.Next() {
		return nil
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret
}

func (this *Database) GetBusStop(busStop string) *BusStopTable {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()

	query := "SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info WHERE BusStopCode = ?"
	rows, err := db.Query(query, busStop)
	if err != nil {
		log.Panic(err)
	}

	// I only expect 1 row
	if !rows.Next() {
		return nil
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret
}
