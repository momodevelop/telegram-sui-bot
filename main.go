package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"telegram-sui-bot/pkg/data"
	"telegram-sui-bot/pkg/director"
	"telegram-sui-bot/pkg/lta"
	"telegram-sui-bot/pkg/repos"
	"telegram-sui-bot/pkg/scenes"
	"telegram-sui-bot/pkg/telegramBot"
)

type Config struct {
	TelegramToken string `json:"telegramToken"`
	LtaToken      string `json:"ltaToken"`
	DbPath        string `json:"dbPath"`
}

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

func main() {
	config := initConfig()
	db := data.NewMysqlDatabase()
	err := db.Open("database.db")
	defer db.Close()
	if err != nil {
		log.Panicf("Cannot initialize database\n%s", err.Error())
	}
	lta := lta.New(config.LtaToken)

	repo := repos.NewRepoBusStops(db, lta)

	bot := telegramBot.Bot{
		Token: config.TelegramToken,
	}

	// stage init
	director := director.New()
	director.Add(
		scenes.NewSceneMain(),
		scenes.NewSceneBus(repo),
	)
	director.SetDefaultScene("Main")
	bot.AddCallbackQueryHandler(scenes.NewBusRefreshCallbackQuery(repo))
	bot.AddMessageHandler(director)
	bot.Run()
}
