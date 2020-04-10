package scenes

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"telegram_go_sui_bot/pkg/database"
	"telegram_go_sui_bot/pkg/director"
	"telegram_go_sui_bot/pkg/lta"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneBusCallbackMiddleware struct {
	busAPI *lta.API
	db     *database.Database
}

type BusRefreshCallbackData struct {
	Cmd       string `json:"cmd"`
	BusStop   string `json:"busStop"`
	TimeStamp int    `json:"timeStamp"`
}

type SceneBus struct {
	busAPI *lta.API
	db     *database.Database
}

func NewSceneBus(busAPI *lta.API, db *database.Database) *SceneBus {
	return &SceneBus{
		busAPI: busAPI,
		db:     db,
	}
}

func (this *SceneBus) Name() string {
	return "Bus"
}

func (this *SceneBus) Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	message := "I can help you check for the buses at your bus stop...!\nJust key in the bus stop number or send me your location and I'll try to find the timings ASAP!（｀・ω・´）\n"
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = getSceneBusKeyboard()
	bot.Send(msg)
}

func (this *SceneBus) Process(session *director.Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	defer recovery(bot, update)

	msg := update.Message.Text
	if len(msg) > 0 {
		if msg == "/exit" {
			session.ChangeScene("Main")
			return
		}

		rex, err := regexp.Compile(`\d+`)
		errCheck("Regex is problemetic", err)

		matches := rex.FindAllString(msg, -1)

		if len(matches) > 0 {
			busStopCode, err := strconv.Atoi(matches[0])
			if err != nil || len(matches[0]) > 5 {
				bot.Send(TelegramAPI.NewMessage(update.Message.Chat.ID, "Invalid bus stop code! Try again..."))
				return
			}
			busStopCodeStr := busStop2Str(busStopCode)
			busStop := this.db.GetBusStop(padStart(busStopCodeStr, "0", 5))
			if busStop == nil {
				bot.Send(TelegramAPI.NewMessage(update.Message.Chat.ID, "That bus stop does not exist!"))
				return
			}

			busArrival := this.busAPI.CallBusArrivalv2(busStopCodeStr, "")
			if len(busArrival.Services) == 0 {
				bot.Send(TelegramAPI.NewMessage(update.Message.Chat.ID, "There are no more buses coming...Maybe you should hail the cab? ^^a"))
				return
			}

			message := createBusETAMessage(busArrival, busStop)
			msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
			msg.ParseMode = "markdown"
			msg.ReplyMarkup = getBusSceneInlineRefreshKeyboard(busStopCodeStr, time.Now().Nanosecond())
			bot.Send(msg)

			return
		}

		this.Greet(bot, update)

	}

}

func getSceneBusKeyboard() TelegramAPI.ReplyKeyboardMarkup {
	return TelegramAPI.NewReplyKeyboard(
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.KeyboardButton{
				Text:            "Send Location!",
				RequestLocation: true,
			},
		),
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/poke"),
		),
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/exit"),
		),
	)
}

// helper functions
func createBusETAMessage(arrival *lta.BusArrivalv2, stop *database.BusStopTable) string {
	currentDate := time.Now()
	currentDateStr := fmt.Sprintf("_Last Updated: %02d-%02d-%02d %02d:%02d:%02d_", currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(), currentDate.Minute(), currentDate.Second())

	if len(arrival.Services) == 0 {
		return fmt.Sprintf("There are no more buses coming...Maybe you should hail the cab? ^^a\n %s", currentDateStr)
	}

	resultArr := make([]string, 0, len(arrival.Services))
	resultArr = append(resultArr, fmt.Sprintf("*%s, %s*```\n", arrival.BusStopCode, stop.Description))

	for _, e := range arrival.Services {
		// format:
		// <bus no>\t:\t<time>\t|\t<time>
		// 131\t:\t1min\t|\t2minw
		// 131	:	1min	|	2min
		strService := padEnd(e.ServiceNo, " ", 5)
		var strBus1 string
		var strBus2 string
		if e.NextBus.EstimatedArrival == "" {
			strBus1 = "-"
		} else {
			strBus1 = padStart(strconv.Itoa(nextBusETA(e.NextBus.EstimatedArrival)), " ", 3)
		}

		if e.NextBus2.EstimatedArrival == "" {
			strBus2 = "-"
		} else {
			strBus2 = padStart(strconv.Itoa(nextBusETA(e.NextBus2.EstimatedArrival)), " ", 3)
		}
		resultArr = append(resultArr, fmt.Sprintf("%s:%s mins | %s mins\n", strService, strBus1, strBus2))
	}

	resultArr = append(resultArr, "```")
	resultArr = append(resultArr, currentDateStr)

	return strings.Join(resultArr, "")
}

func busStop2Str(busStop int) string {
	return fmt.Sprintf("%05d", busStop)
}

// Returns the number of minutes for the next bus to come
func nextBusETA(estimatedTimeArr string) int {

	//"EstimatedArrival": "2017-06-05T14:57:09+08:00"
	/*
		let time:string = nextBusthis.EstimatedArrival;
		let t_index:number = time.indexOf("T");
		let plus_index:number = time.indexOf("+");
		time = time.substr(t_index + 1, plus_index);
	*/

	etaDate, err := time.Parse("2006-01-02T15:04:05-07:00", estimatedTimeArr)
	errCheck("[SceneBus][nextBusETA]", err)

	now := time.Now()

	diff := etaDate.Sub(now)
	mins := diff.Minutes()

	if mins > 0 {
		return int(mins)
	}

	return 0
}

func getBusSceneInlineRefreshKeyboard(busStop string, timeStamp int) TelegramAPI.InlineKeyboardMarkup {

	bytes, err := json.Marshal(BusRefreshCallbackData{
		Cmd:       "refresh",
		BusStop:   busStop,
		TimeStamp: timeStamp,
	})
	errCheck("[getBusSceneInlineRefreshKeyboard] Problems converting callback data to string", err)

	return TelegramAPI.NewInlineKeyboardMarkup(
		TelegramAPI.NewInlineKeyboardRow(
			TelegramAPI.NewInlineKeyboardButtonData("Refresh", string(bytes)),
		),
	)

}

func NewSceneBusCallbackMiddleware(busAPI *lta.API, db *database.Database) *SceneBusCallbackMiddleware {
	return &SceneBusCallbackMiddleware{
		busAPI: busAPI,
		db:     db,
	}
}

func (this *SceneBusCallbackMiddleware) Process(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {

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
			return
		}

		if callbackData.Cmd == "refresh" {
			busStop := this.db.GetBusStop(callbackData.BusStop)
			if busStop == nil {
				bot.Send(TelegramAPI.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"That bus stop does not exist!",
				))
				return
			}

			busArrival := this.busAPI.CallBusArrivalv2(callbackData.BusStop, "")
			if len(busArrival.Services) == 0 {
				bot.Send(TelegramAPI.NewEditMessageText(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
					"There are no more buses coming...Maybe you should hail the cab? ^^a",
				))
				return
			}

			reply := createBusETAMessage(busArrival, busStop)
			msg := TelegramAPI.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, reply)
			msg.ParseMode = "markdown"
			keyboard := getBusSceneInlineRefreshKeyboard(callbackData.BusStop, time.Now().Nanosecond())
			msg.ReplyMarkup = &keyboard
			bot.Send(msg)

		}

	}

}
