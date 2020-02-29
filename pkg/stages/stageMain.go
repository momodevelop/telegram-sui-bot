package stages

import (
	stgmgr "telegram_go_sui_bot/pkg/stageManager"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type StageMain struct{}

func (obj *StageMain) Name() string {
	return "Main"
}

func (obj *StageMain) Greet(session *stgmgr.Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {

}

func (obj *StageMain) Process(session *stgmgr.Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Main!")
	msg.ReplyToMessageID = update.Message.MessageID

	bot.Send(msg)
}
