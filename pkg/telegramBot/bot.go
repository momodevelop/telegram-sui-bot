package telegramBot

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type IMessageHandler interface {
	Process(*TelegramAPI.BotAPI, *TelegramAPI.Message) bool
}

type ICallbackQueryHandler interface {
	Process(*TelegramAPI.BotAPI, *TelegramAPI.CallbackQuery) bool
}

type Bot struct {
	Token                 string
	messageHandlers       []IMessageHandler
	callbackQueryHandlers []ICallbackQueryHandler
}

func (this *Bot) AddMessageHandler(handler ...IMessageHandler) {
	this.messageHandlers = append(this.messageHandlers, handler...)
}

func (this *Bot) AddCallbackQueryHandler(handler ...ICallbackQueryHandler) {
	this.callbackQueryHandlers = append(this.callbackQueryHandlers, handler...)
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
			for _, handler := range this.callbackQueryHandlers {
				if !handler.Process(bot, update.CallbackQuery) {
					break
				}
			}
		} else if update.Message != nil {
			for _, handler := range this.messageHandlers {
				if !handler.Process(bot, update.Message) {
					break
				}
			}
		}
	}
}
