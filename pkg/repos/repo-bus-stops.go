package repos

import (
	"fmt"
	"log"
	"strings"
	"telegram-sui-bot/pkg/data"
	"telegram-sui-bot/pkg/lta"

	_ "github.com/mattn/go-sqlite3"
)

var dbType string = "sqlite3"

type RepoBusStops struct {
	Db  *data.MysqlDatabase
	Lta *lta.API
}

func NewRepoBusStops(Db *data.MysqlDatabase, Lta *lta.API) *RepoBusStops {
	ret := &RepoBusStops{
		Db:  Db,
		Lta: Lta,
	}
	return ret
}

func (this *RepoBusStops) GetBusStopArrivalsByNearestLocation(latitude float64, longitude float64) (*BusStopTable, error) {
	query := fmt.Sprintf("SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info ORDER BY (Latitude - ?) * (Latitude - ?) + (Longitude - ?) * (Longitude - ?)")
	rows, err := this.Db.Query(query, latitude, latitude, longitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][GetBusStopArrivalsByNearestLocation] Problem with query\n%s", err.Error())
	}

	// I only expect 1 row
	if !rows.Next() {
		return nil, fmt.Errorf("[RepoBusStops][GetBusStopArrivalsByNearestLocation] There are no rows??")
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret, nil
}

func (this *RepoBusStops) GetBusStop(busStop string) (*BusStopTable, error) {
	query := "SELECT BusStopCode, RoadName, Description, Latitude, Longitude FROM bus_stop_info WHERE BusStopCode = ?"
	rows, err := this.Db.Query(query, busStop)
	if err != nil {
		return nil, fmt.Errorf("[RepoBusStops][GetBusStop] Problem with query\n%s", err.Error())
	}
	// I only expect 1 row
	if !rows.Next() {
		return nil, fmt.Errorf("[RepoBusStops][GetBusStop] There are no rows??")
	}

	var ret BusStopTable
	rows.Scan(&ret.BusStopCode, &ret.RoadName, &ret.Description, &ret.Latitude, &ret.Longitude)
	return &ret, nil
}

func (this *RepoBusStops) GetBusStopArrivals(busStop string) (*lta.BusArrivalv2, error) {
	path := "ltaodataservice/BusArrivalv2?BusStopCode=" + busStop

	var ret *lta.BusArrivalv2
	err := this.Lta.Call(path, &ret)
	if err != nil {
		return nil, fmt.Errorf("[LTA][GetBusArrival] Something went wrong\n%s", err.Error())
	}
	return ret, nil
}

func (this *RepoBusStops) SyncBusStops(busStop string) error {
	// Retreive bus stops from LTA
	var err error
	log.Println("[LTA][SyncBusStops] Retrieving bus stops from API!")
	busStops := make([]BusStopTable, 0)
	skip := 0
	totalStops := 0
	for {
		busStopResponse, err := this.Lta.CallBusStops(skip)
		if err != nil {
			return fmt.Errorf("[LTA][SyncBusStops] Cannot call bus stops\n%s", err.Error())
		}
		if busStopResponse != nil && len(busStopResponse.Value) > 0 {
			totalStops += len(busStopResponse.Value)
			skip += 500
			log.Printf("[LTA][SyncBusStops] %d stops...", totalStops)
			for _, e := range busStopResponse.Value {
				var table BusStopTable
				table.BusStopCode = e.BusStopCode
				table.Description = e.Description
				table.Latitude = e.Latitude
				table.Longitude = e.Longitude
				table.RoadName = e.RoadName
				busStops = append(busStops, table)
			}
		} else {
			break
		}
	}

	log.Printf("[LTA][SyncBusStops] %d entries inserted!\n", totalStops)

	// Insert into database
	transaction, err := this.Db.Begin()
	if err != nil {
		return fmt.Errorf("[LTA][SyncBusStops] Problems preparing transaction\n%s", err.Error())
	}

	defer func() error {
		if err != nil {
			err = transaction.Rollback()
			if err != nil {
				return fmt.Errorf("[LTA][SyncBusStops] Problems with rollback\n%s", err.Error())
			}
		}
		return nil
	}()

	_, err = transaction.Exec("DROP TABLE IF EXISTS bus_stop_info")
	if err != nil {
		return fmt.Errorf("[LTA][SyncBusStops] Error with dropping table\n%s", err.Error())
	}

	_, err = transaction.Exec("CREATE TABLE IF NOT EXISTS bus_stop_info(BusStopCode TEXT PRIMARY KEY, RoadName TEXT, Description TEXT, Latitude float(24), Longitude float(24))")
	if err != nil {
		return fmt.Errorf("[LTA][SyncBusStops] Error with create table\n%s", err.Error())
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
				return fmt.Errorf("[LTA][SyncBusStops] Error with insert into\n%s", err.Error())
			}
			queryInsertIntoStr = queryInsertIntoStr[:0]
			queryInsertIntoArgs = queryInsertIntoArgs[:0]
			batchCounter = 0
		}
	}

	// Commit
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("[LTA][SyncBusStops] Error committing transaction\n%s", err.Error())
	}

	return nil
}
