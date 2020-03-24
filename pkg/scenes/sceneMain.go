package scenes

import (
	"fmt"
	"math/rand"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneMain struct{}

func NewSceneMain() *SceneMain {
	return &SceneMain{}
}

func (obj *SceneMain) Name() string {
	return "Main"
}

func (obj *SceneMain) Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	message := "Erm...Hi? I'm Tachibana Sui, your humble...er...bot.\nI can help you do a few things, just give me one of the commands:\n---\n/bus to get bus ETA\n/food if you want me to help you decide what to eat"
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = obj.getKeyboard()
	bot.Send(msg)
}

func (obj *SceneMain) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) (bool, string) {
	switch update.Message.Text {
	case "/bus":
		return true, "Bus"
	case "/food":
		msg := TelegramAPI.NewMessage(update.Message.Chat.ID, fmt.Sprintf("You should have *%s*!", getRandomFoodRecommandation()))
		msg.ParseMode = "markdown"
		msg.ReplyMarkup = obj.getKeyboard()
		bot.Send(msg)
	default:
		obj.Greet(bot, update)
	}
	return false, ""
}

func (obj *SceneMain) getKeyboard() TelegramAPI.ReplyKeyboardMarkup {
	return TelegramAPI.NewReplyKeyboard(
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/bus"),
		),
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/food"),
		),
	)
}

func getRandomFoodRecommandation() string {
	recommendations := []string{
		//general
		"something with curry",
		"something soupy",
		"something with rice",
		"something with bread",
		"something with noodles",

		// specific
		"udon",
		"soba",
		"sushi",
		"ramen",
		"pasta",
		"pizza",
		"burger",
		"wrap",
		"sandwich",

		// meat
		"something with beef",
		"something with chicken",
		"something with pork",
		"something with fish",
		"something with meat",
		"something with vegetables",

		//cultural
		"Indian",
		"Western",
		"Japanese",
		"Korean",
		"Chinese",
		"Italian",
		"Mexican",
		"Turkish",
	}

	return recommendations[rand.Intn(len(recommendations))]
}
