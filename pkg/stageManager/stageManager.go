package stageManager

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Session struct {
	Stage string
}
type IStage interface {
	Name() string
	Greet(session *Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update)
	Process(session *Session, bot *tgbotapi.BotAPI, update *tgbotapi.Update)
}

type user_t int
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

func (this *Manager) Process(userID int, bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Check if session exists
	session, ok := this.sessions[userID]
	if !ok {
		this.sessions[userID] = &Session{Stage: "Main"}
		session, ok = this.sessions[userID]
		if !ok {
			log.Panicf("Cannot create session for %d", userID)
		}
	}

	// redirect based on session
	stage, ok := this.stages[session.Stage]
	if !ok {
		log.Panicf("Invalid stage: %s", session.Stage)
	}
	stage.Process(session, bot, update)
}
