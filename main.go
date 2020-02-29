package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	stgmgr "telegram_go_sui_bot/pkg/stageManager"
	"telegram_go_sui_bot/pkg/stages"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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

	// stage init
	stageMgr := stgmgr.New()
	stageMgr.Add(
		&stages.StageMain{},
		&stages.StageBus{},
	)

	bot, err := tgbotapi.NewBotAPI(config["telegramToken"])
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		user := update.Message.From

		if update.Message == nil { // ignore any non-Message Updates
			return
		}
		log.Printf("Message is from %d", user.ID)
		stageMgr.Process(user.ID, bot, &update)
	}
}
