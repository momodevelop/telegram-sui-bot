package scenes

import (
	"encoding/json"
	"log"
	"telegram_go_sui_bot/pkg/database"
	"telegram_go_sui_bot/pkg/lta"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BusRefreshCallbackQuery struct {
	busAPI *lta.API
	db     *database.Database
}

type BusRefreshCallbackData struct {
	Cmd       string `json:"cmd"`
	BusStop   string `json:"busStop"`
	TimeStamp int    `json:"timeStamp"`
}

func NewBusRefreshCallbackQuery(busAPI *lta.API, db *database.Database) *BusRefreshCallbackQuery {
	return &BusRefreshCallbackQuery{
		busAPI: busAPI,
		db:     db,
	}
}

func (this *BusRefreshCallbackQuery) Process(bot *TelegramAPI.BotAPI, callbackQuery *TelegramAPI.CallbackQuery) bool {
	defer func() {
		if r := recover(); r != nil {
			msg := TelegramAPI.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, "Oops, something went wrong ><")
			bot.Send(msg)

			log.Println("Stop panicking: ", r)
		}
	}()

	// Check if CallbackQuery is parsable to BusRefreshCallbackData
	var callbackData BusRefreshCallbackData
	err := json.Unmarshal([]byte(callbackQuery.Data), &callbackData)
	if err != nil {
		return true
	}

	if callbackData.Cmd == "refresh" {
		busStop := this.db.GetBusStop(callbackData.BusStop)
		if busStop == nil {
			bot.Send(TelegramAPI.NewEditMessageText(
				callbackQuery.Message.Chat.ID,
				callbackQuery.Message.MessageID,
				"That bus stop does not exist!",
			))
			return true
		}

		busArrival := this.busAPI.CallBusArrivalv2(callbackData.BusStop, "")
		if len(busArrival.Services) == 0 {
			bot.Send(TelegramAPI.NewEditMessageText(
				callbackQuery.Message.Chat.ID,
				callbackQuery.Message.MessageID,
				"There are no more buses coming...Maybe you should hail the cab? ^^a",
			))
			return true
		}

		reply := createBusETAMessage(busArrival, busStop)
		msg := TelegramAPI.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, reply)
		msg.ParseMode = "markdown"
		keyboard := getBusSceneInlineRefreshKeyboard(callbackData.BusStop, time.Now().Nanosecond())
		msg.ReplyMarkup = &keyboard
		bot.Send(msg)

	}

	return false

}
