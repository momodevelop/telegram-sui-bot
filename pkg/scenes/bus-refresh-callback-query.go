package scenes

import (
	"encoding/json"
	"fmt"
	"log"
	"telegram-sui-bot/pkg/repos"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type BusRefreshCallbackQuery struct {
	RepoBusStops *repos.RepoBusStops
}

type BusRefreshCallbackData struct {
	Cmd       string `json:"cmd"`
	BusStop   string `json:"busStop"`
	TimeStamp int    `json:"timeStamp"`
}

func NewBusRefreshCallbackQuery(RepoBusStops *repos.RepoBusStops) *BusRefreshCallbackQuery {
	return &BusRefreshCallbackQuery{
		RepoBusStops: RepoBusStops,
	}
}

func (this *BusRefreshCallbackQuery) Process(bot *TelegramAPI.BotAPI, callbackQuery *TelegramAPI.CallbackQuery) bool {
	// Check if CallbackQuery is parsable to BusRefreshCallbackData
	var callbackData BusRefreshCallbackData
	err := json.Unmarshal([]byte(callbackQuery.Data), &callbackData)
	if err != nil {
		log.Printf("BusRefreshCallbackQuery][Process] Something weng wrong parsing json\n%s", err.Error())
		return true
	}

	if callbackData.Cmd == "refresh" {
		busStop, err := this.RepoBusStops.GetBusStop(callbackData.BusStop)
		if err != nil {
			log.Printf("BusRefreshCallbackQuery][Process] Something went wrong getting bus stop\n%s", err.Error())
			return false
		}
		if busStop == nil {
			bot.Send(TelegramAPI.NewEditMessageText(
				callbackQuery.Message.Chat.ID,
				callbackQuery.Message.MessageID,
				"That bus stop does not exist!",
			))
			return true
		}

		busArrival, err := this.RepoBusStops.GetBusStopArrivals(callbackData.BusStop)
		if err != nil {
			log.Printf("[BusRefreshCallbackQuery][Process] Something went wrong getting bus arrival\n%s", err.Error())
			return false
		}

		keyboard, err := getBusSceneInlineRefreshKeyboard(callbackData.BusStop, time.Now().Nanosecond())
		if err != nil {
			log.Printf("[BusRefreshCallbackQuery][Process] Something went wrong with creating refresh keyboard\n%s", err.Error())
			return false
		}

		msg := TelegramAPI.NewEditMessageText(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID, "")
		msg.ReplyMarkup = keyboard
		msg.ParseMode = "markdown"

		if len(busArrival.Services) == 0 {
			msg.Text = fmt.Sprintf("There are no more buses coming for ***%s***...Maybe you should hail the cab? ^^a", busStop.RoadName)
			msg.ParseMode = "markdown"
			bot.Send(msg)

		} else {
			reply, err := createBusETAMessage(busArrival, busStop)
			if err != nil {
				log.Printf("[BusRefreshCallbackQuery][Process] Something went wrong with creating a bus ETA message\n%s", err.Error())
				return false
			}
			msg.Text = *reply
		}
		msg.ParseMode = "markdown"
		bot.Send(msg)
		return true

	}

	return false

}
