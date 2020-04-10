package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"telegram_go_sui_bot/pkg/database"
	"telegram_go_sui_bot/pkg/director"
	"telegram_go_sui_bot/pkg/lta"
	"telegram_go_sui_bot/pkg/scenes"
	"telegram_go_sui_bot/pkg/telegramBot"
)

type Config map[string]string

func initConfig() Config {
	log.Println("Initializing config...")
	var config Config
	ex, err := os.Executable()
	if err != nil {
		log.Panic(err)
	}
	exPath := filepath.Dir(ex)

	file, err := os.Open(exPath + "/config.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Panic(err)
	}
	return config
}

func syncBusStopsFromApiToDb(lta *lta.API, db *database.Database) {
	log.Println("[syncBusStopsFromApiToDb] Retrieving bus stops from API!")
	busStops := make([]database.BusStopTable, 0)
	skip := 0
	totalStops := 0
	for {
		busStopResponse := lta.CallBusStops(skip)
		if busStopResponse != nil && len(busStopResponse.Value) > 0 {
			totalStops += len(busStopResponse.Value)
			skip += 500

			for _, e := range busStopResponse.Value {
				var table database.BusStopTable
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

	log.Printf("[syncBusStopsFromApiToDb] %d entries inserted!\n", totalStops)
	db.RefreshBusStopTable(busStops)
}

func main() {
	config := initConfig()
	db := database.New("database.db")
	lta := lta.New(config["ltaToken"])

	bot := telegramBot.Bot{
		Token: config["telegramToken"],
	}

	syncBusStopsFromApiToDb(lta, db)

	// stage init
	director := director.New()
	director.Add(
		scenes.NewSceneMain(),
		scenes.NewSceneBus(lta, db),
	)
	director.SetDefaultScene("Main")
	bot.AddMiddleware(scenes.NewSceneBusCallbackMiddleware(lta, db))
	bot.AddMiddleware(director)
	bot.Run()

	log.Println("Exiting")
}
