package scenes

import (
	"encoding/json"
	"log"
	"telegram_go_sui_bot/pkg/database"
	"telegram_go_sui_bot/pkg/lta"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

func NewSceneBusCallbackMiddleware(busAPI *lta.API, db *database.Database) *SceneBusCallbackMiddleware {
	return &SceneBusCallbackMiddleware{
		busAPI: busAPI,
		db:     db,
	}
}

func (this *SceneBusCallbackMiddleware) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) bool {

	if update.CallbackQuery != nil {
		defer func(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
			if r := recover(); r != nil {
				msg := TelegramAPI.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, "Oops, something went wrong ><")
				bot.Send(msg)

				log.Println("Stop panicking: ", r)
			}
		}(bot, update)

		// Check if CallbackQuery is parsable to BusRefreshCallbackData
		var callbackData BusRefreshCallbackData
		err := json.Unmarshal([]byte(update.CallbackQuery.Data), &callbackData)
		if err != nil {
			return true
		}

		if callbackData.Cmd == "refresh" {
			busStop := this.db.GetBusStop(callbackData.BusStop)
			if busStop == nil {
				bot.Send(TelegramAPI.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"That bus stop does not exist!",
				))
				return true
			}

			busArrival := this.busAPI.CallBusArrivalv2(callbackData.BusStop, "")
			if len(busArrival.Services) == 0 {
				bot.Send(TelegramAPI.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"There are no more buses coming...Maybe you should hail the cab? ^^a",
				))
				return true
			}

			reply := createBusETAMessage(busArrival, busStop)
			msg := TelegramAPI.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, reply)
			msg.ParseMode = "markdown"
			keyboard := getBusSceneInlineRefreshKeyboard(callbackData.BusStop, time.Now().Nanosecond())
			msg.ReplyMarkup = &keyboard
			bot.Send(msg)

		}

	}

	return true

}
