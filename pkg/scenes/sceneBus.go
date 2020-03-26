package scenes

import (
	Director "telegram_go_sui_bot/pkg/director"
	LtaAPI "telegram_go_sui_bot/pkg/lta"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneBus struct {
	busAPI *LtaAPI.API
}

func NewSceneBus(landTransportDataMallToken string) *SceneBus {
	return &SceneBus{
		busAPI: LtaAPI.New(landTransportDataMallToken),
	}
}

func (obj *SceneBus) Name() string {
	return "Bus"
}

func (obj *SceneBus) Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, "Bus Greet!")
	bot.Send(msg)
}

func (obj *SceneBus) Process(session *Director.Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {

	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, "Bus!")
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)

}
