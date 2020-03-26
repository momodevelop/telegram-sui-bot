package director

import TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"

type Session struct {
	scene      string
	hasChanged bool
}

type IScene interface {
	Name() string
	Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update)
	Process(session *Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update)
}

func (this *Session) ChangeScene(sceneName string) {
	this.scene = sceneName
	this.hasChanged = true
}