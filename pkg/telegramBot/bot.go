package telegramBot

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type IMiddleware interface {
	Process(*TelegramAPI.BotAPI, *TelegramAPI.Update)
}

type Bot struct {
	Token          string
	middlewareList []IMiddleware
}

func (this *Bot) AddMiddleware(middleware ...IMiddleware) {
	this.middlewareList = append(this.middlewareList, middleware...)
}

func (this *Bot) Run() {
	bot, err := TelegramAPI.NewBotAPI(this.Token)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := TelegramAPI.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		for _, middleware := range this.middlewareList {
			middleware.Process(bot, &update)
		}
	}
}
