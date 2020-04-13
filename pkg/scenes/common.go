package scenes

import (
	"log"
	"strings"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

func recovery(bot *TelegramAPI.BotAPI, message *TelegramAPI.Message) {
	if r := recover(); r != nil {
		msg := TelegramAPI.NewMessage(message.Chat.ID, "Oops, something went wrong ><")
		bot.Send(msg)

		log.Println("Stop panicking: ", r)
	}
}

func padStart(str string, item string, count int) string {
	padAmount := count - len(str)
	if padAmount > 0 {
		return strings.Repeat(item, padAmount) + str
	}
	return str

}

func padEnd(str string, item string, count int) string {
	padAmount := count - len(str)
	if padAmount > 0 {
		return str + strings.Repeat(item, padAmount)
	}
	return str
}
