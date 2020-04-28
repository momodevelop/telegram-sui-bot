package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var dbType string = "sqlite3"

type RepoBusStops struct {
	dbPath string
}

func New(dbPath string) (*RepoBusStops, error) {
	log.Println("Checking database...")
	db, err := sql.Open(dbType, dbPath)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][New] Cannot open DB\n%s", err.Error())
	}
	defer db.Close()

	ret := RepoBusStops{
		dbPath: dbPath,
	}

	return &ret, nil
}

func (this *RepoBusStops) RefreshBusStopTable(busStops []BusStopTable) error {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Cannot open DB\n%s", err.Error())
	}
	defer db.Close()

	queryDropTable := "DROP TABLE IF EXISTS bus_stop_info"
	queryCreateTable := "CREATE TABLE IF NOT EXISTS bus_stop_info(BusStopCode TEXT PRIMARY KEY, RoadName TEXT, Description TEXT, Latitude float(24), Longitude float(24))"

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Problems preparing transaction\n%s", err.Error())
	}

	defer func() error {
		if err != nil {
			err = transaction.Rollback()
			if err != nil {
				return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Problems with rollback\n%s", err.Error())
			}
		}
		return nil
	}()

	_, err = transaction.Exec(queryDropTable)
	if err != nil {
		return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Error with dropping table\n%s", err.Error())
	}

	_, err = transaction.Exec(queryCreateTable)
	if err != nil {
		return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Error with create table\n%s", err.Error())
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
				return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Error with insert into\n%s", err.Error())
			}
			queryInsertIntoStr = queryInsertIntoStr[:0]
			queryInsertIntoArgs = queryInsertIntoArgs[:0]
			batchCounter = 0
		}
	}

	// Commit
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("[RepoBusStops][ResetBusStopTable] Error committing transaction\n%s", err.Error())
	}

	return nil
}

func (this *RepoBusStops) GetBusStopByNearestLocation(latitude float64, longitude float64) (*BusStopTable, error) {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][ResetBusStopTable] Cannot open DB\n%s", err.Error())
	}
	defer db.Close()

	query := fmt.Sprintf("SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info ORDER BY (Latitude - ?) * (Latitude - ?) + (Longitude - ?) * (Longitude - ?)")
	rows, err := db.Query(query, latitude, latitude, longitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][DoesBusStopExist] Problem with query\n%s", err.Error())
	}

	// I only expect 1 row
	if !rows.Next() {
		return nil, fmt.Errorf("[RepoBusStops][DoesBusStopExist] There are no rows??")
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret, nil
}

func (this *RepoBusStops) GetBusStop(busStop string) (*BusStopTable, error) {
	db, err := sql.Open(dbType, this.dbPath)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][ResetBusStopTable] Cannot open DB\n%s", err.Error())
	}
	defer db.Close()

	query := "SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info WHERE BusStopCode = ?"
	rows, err := db.Query(query, busStop)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][DoesBusStopExist] Problem with query\n%s", err.Error())
	}
	// I only expect 1 row
	if !rows.Next() {
		return nil, fmt.Errorf("[RepoBusStops][DoesBusStopExist] There are no rows??")
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret, nil
}
