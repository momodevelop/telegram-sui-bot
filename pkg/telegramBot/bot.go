package telegramBot

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type IMessageMiddleware interface {
	Process(*TelegramAPI.BotAPI, *TelegramAPI.Message) bool
}

type ICallbackQueryMiddleware interface {
	Process(*TelegramAPI.BotAPI, *TelegramAPI.CallbackQuery) bool
}

type Bot struct {
	Token                    string
	messageMiddlewares       []IMessageMiddleware
	callbackQueryMiddlewares []ICallbackQueryMiddleware
}

func (this *Bot) AddMessageMiddleware(middleware ...IMessageMiddleware) {
	this.messageMiddlewares = append(this.messageMiddlewares, middleware...)
}

func (this *Bot) AddCallbackQueryMiddleware(middleware ...ICallbackQueryMiddleware) {
	this.callbackQueryMiddlewares = append(this.callbackQueryMiddlewares, middleware...)
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

		if update.CallbackQuery != nil {
			for _, middleware := range this.callbackQueryMiddlewares {
				if !middleware.Process(bot, update.CallbackQuery) {
					break
				}
			}
		} else if update.Message != nil {
			for _, middleware := range this.messageMiddlewares {
				if !middleware.Process(bot, update.Message) {
					break
				}
			}
		}
	}
}
