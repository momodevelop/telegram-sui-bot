package stages

import (
	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type StageBus struct {
}

func NewStageBus() *StageBus {
	return &StageBus{}
}

func (obj *StageBus) Name() string {
	return "Bus"
}

func (obj *StageBus) Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, "Bus Greet!")
	bot.Send(msg)
}

func (obj *StageBus) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) (bool, string) {

	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, "Bus!")
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)

	return false, ""
}
