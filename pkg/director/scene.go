package director

import TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"

type Session struct {
	scene      string
	hasChanged bool
}

type IScene interface {
	Name() string
	Process(session *Session, bot *TelegramAPI.BotAPI, message *TelegramAPI.Message)
}

func (this *Session) ChangeScene(sceneName string) {
	this.scene = sceneName
	this.hasChanged = true
}
