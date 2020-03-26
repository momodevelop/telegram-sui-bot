package director

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Manager struct {
	sessions         map[int]*Session
	scenes           map[string]IScene
	defaultSceneName string
}

func New() *Manager {
	return &Manager{
		sessions: make(map[int]*Session),
		scenes:   make(map[string]IScene),
	}
}

func (this *Manager) SetDefaultScene(sceneName string) {
	_, ok := this.scenes[sceneName]
	if !ok {
		log.Panicf("Scene does not exist: %s", sceneName)
	}
	this.defaultSceneName = sceneName
}

func (this *Manager) Add(scenes ...IScene) {
	for _, scene := range scenes {
		log.Println("Adding Scene: " + scene.Name())
		this.scenes[scene.Name()] = scene
	}

}

func (this *Manager) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	if len(this.defaultSceneName) == 0 {
		log.Panicf("Default scene does not exist! Please set with SetDefaultScene()")
	}
	user := update.Message.From

	// Check if session exists. If it does not, create new session
	session, oldUser := this.sessions[user.ID]
	if !oldUser {
		var ok bool
		this.sessions[user.ID] = &Session{scene: this.defaultSceneName}
		session, ok = this.sessions[user.ID]
		if !ok {
			log.Panicf("Cannot create session for %d", user.ID)
		}
	}

	// redirect based on session
	scene, ok := this.scenes[session.scene]
	if !ok {
		log.Panicf("Invalid Scene: %s", session.scene)
	}
	if !oldUser {
		scene.Greet(bot, update)
	} else {
		scene.Process(session, bot, update)
		if session.hasChanged {
			sceneToChange, ok := this.scenes[session.scene]
			if !ok {
				log.Panicf("Invalid Scene to change: %s", session.scene)
			}
			session.scene = sceneToChange.Name()
			sceneToChange.Greet(bot, update)
		}

	}

}
