package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"telegram_go_sui_bot/pkg/stageManager"
	"telegram_go_sui_bot/pkg/stages"
	"telegram_go_sui_bot/pkg/telegramBot"
)

type config_t map[string]string

func initConfig() config_t {
	log.Println("Initializing config...")
	var config config_t
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

func main() {
	config := initConfig()
	bot := telegramBot.Bot{
		Token: config["telegramToken"],
	}

	// stage init
	stageMgr := stageManager.New()
	stageMgr.Add(
		&stages.StageMain{},
		&stages.StageBus{},
	)
	bot.AddMiddleware(stageMgr)
	bot.Run()

}
