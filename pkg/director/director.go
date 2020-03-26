package director

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Session struct {
	Scene string
}
type IScene interface {
	Name() string
	Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update)
	Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) (bool, string)
}

type Manager struct {
	sessions map[int]*Session
	scenes   map[string]IScene
}

func New() *Manager {
	return &Manager{
		sessions: make(map[int]*Session),
		scenes:   make(map[string]IScene),
	}
}

func (this *Manager) Add(scenes ...IScene) {
	for _, Scene := range scenes {
		log.Println("Adding Scene: " + Scene.Name())
		this.scenes[Scene.Name()] = Scene
	}

}

func (this *Manager) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	user := update.Message.From

	// Check if session exists. If it does not, create new session
	session, oldUser := this.sessions[user.ID]
	if !oldUser {
		var ok bool
		this.sessions[user.ID] = &Session{Scene: "Main"}
		session, ok = this.sessions[user.ID]
		if !ok {
			log.Panicf("Cannot create session for %d", user.ID)
		}
	}

	// redirect based on session
	Scene, ok := this.scenes[session.Scene]
	if !ok {
		log.Panicf("Invalid Scene: %s", session.Scene)
	}
	if !oldUser {
		Scene.Greet(bot, update)
	} else {
		isChangeScene, SceneToChangeStr := Scene.Process(bot, update)
		if isChangeScene {
			SceneToChange, ok := this.scenes[SceneToChangeStr]
			if !ok {
				log.Panicf("Invalid Scene to change: %s", session.Scene)
			}
			session.Scene = SceneToChange.Name()
			SceneToChange.Greet(bot, update)
		}

	}

}
