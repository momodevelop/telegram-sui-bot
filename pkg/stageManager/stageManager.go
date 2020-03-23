package stageManager

import (
	"log"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Session struct {
	Stage string
}
type IStage interface {
	Name() string
	Greet(session *Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update)
	Process(session *Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) (bool, string)
}

type Manager struct {
	sessions map[int]*Session
	stages   map[string]IStage
}

func New() *Manager {
	return &Manager{
		sessions: make(map[int]*Session),
		stages:   make(map[string]IStage),
	}
}

func (this *Manager) Add(stages ...IStage) {
	for _, stage := range stages {
		log.Println("Adding Stage: " + stage.Name())
		this.stages[stage.Name()] = stage
	}

}

func (this *Manager) Process(api *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	user := update.Message.From

	// Check if session exists. If it does not, create new session
	session, oldUser := this.sessions[user.ID]
	if !oldUser {
		var ok bool
		this.sessions[user.ID] = &Session{Stage: "Main"}
		session, ok = this.sessions[user.ID]
		if !ok {
			log.Panicf("Cannot create session for %d", user.ID)
		}
	}

	// redirect based on session
	stage, ok := this.stages[session.Stage]
	if !ok {
		log.Panicf("Invalid stage: %s", session.Stage)
	}
	if !oldUser {
		stage.Greet(session, api, update)
	} else {
		isChangeStage, stageToChangeStr := stage.Process(session, api, update)
		if isChangeStage {
			stageToChange, ok := this.stages[stageToChangeStr]
			if !ok {
				log.Panicf("Invalid stage to change: %s", session.Stage)
			}
			session.Stage = stageToChange.Name()
			stageToChange.Greet(session, api, update)
		}

	}

}
