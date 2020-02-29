package stages

import (
	stgmgr "telegram_go_sui_bot/pkg/stageManager"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type StageBus struct{}

func (obj *StageBus) Name() string {
	return "Bus"
}

func (obj *StageBus) Greet(session *stgmgr.Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

}

func (obj *StageBus) Process(session *stgmgr.Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Bus!")
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}
