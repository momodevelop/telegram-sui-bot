package scenes

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"telegram_go_sui_bot/pkg/database"
	"telegram_go_sui_bot/pkg/director"
	"telegram_go_sui_bot/pkg/lta"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

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

func (this *SceneBus) Process(session *director.Session, bot *TelegramAPI.BotAPI, message *TelegramAPI.Message) {
	defer recovery(bot, message)

	if len(message.Text) > 0 {
		msg := message.Text
		if msg == "/exit" {
			session.ChangeScene("Main")
			return
		}

		rex, err := regexp.Compile(`\d+`)
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Regex is problemetic\n%s", err.Error()))
			return
		}

		matches := rex.FindAllString(msg, -1)

		if len(matches) > 0 {
			busStopCode, err := strconv.Atoi(matches[0])
			if err != nil || len(matches[0]) > 5 {
				bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "Invalid bus stop code! Try again..."))
				return
			}
			busStopCodeStr := busStop2Str(busStopCode)
			busStop, err := this.db.GetBusStop(padStart(busStopCodeStr, "0", 5))
			if err != nil {
				tilt(bot, message, fmt.Errorf("[SceneBus][Process] Can't get bus stop\n%s"))
				return
			}
			if busStop == nil {
				bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "That bus stop does not exist!"))
				return
			}

			busArrival := this.busAPI.CallBusArrivalv2(busStopCodeStr, "")
			if len(busArrival.Services) == 0 {
				bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "There are no more buses coming...Maybe you should hail the cab? ^^a"))
				return
			}

			reply, err := createBusETAMessage(busArrival, busStop)
			if err != nil {
				tilt(bot, message, fmt.Errorf("[SceneBus][Process] Can't get bus stop\n%s", err.Error()))
				return
			}
			msg := TelegramAPI.NewMessage(message.Chat.ID, *reply)
			msg.ParseMode = "markdown"

			result, err := getBusSceneInlineRefreshKeyboard(busStop.BusStopCode, time.Now().Nanosecond())
			if err != nil {
				tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong creating refresh keyboard\n%s", err.Error()))
				return
			}
			msg.ReplyMarkup = result
			bot.Send(msg)
		} else {
			reply := "I can help you check for the buses at your bus stop...!\nJust key in the bus stop number or send me your location and I'll try to find the timings ASAP!（｀・ω・´）\n"
			msg := TelegramAPI.NewMessage(message.Chat.ID, reply)
			msg.ReplyMarkup = getSceneBusKeyboard()
			bot.Send(msg)
		}

	} else if message.Location != nil {
		busStop, err := this.db.GetBusStopByNearestLocation(message.Location.Latitude, message.Location.Longitude)
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong getting bus stop by nearest location\n%s", err.Error()))
			return
		}
		if busStop == nil {
			bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "That bus stop does not exist!"))
			return
		}

		busArrival := this.busAPI.CallBusArrivalv2(busStop.BusStopCode, "")
		if len(busArrival.Services) == 0 {
			bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "There are no more buses coming...Maybe you should hail the cab? ^^a"))
			return
		}

		reply, err := createBusETAMessage(busArrival, busStop)
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong getting bus stop by nearest location\n%s", err.Error()))
			return
		}

		msg := TelegramAPI.NewMessage(message.Chat.ID, *reply)
		msg.ParseMode = "markdown"

		result, err := getBusSceneInlineRefreshKeyboard(busStop.BusStopCode, time.Now().Nanosecond())
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong creating refresh keyboard\n%s", err.Error()))
			return
		}
		msg.ReplyMarkup = result
		bot.Send(msg)

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
func createBusETAMessage(arrival *lta.BusArrivalv2, stop *database.BusStopTable) (*string, error) {
	currentDate := time.Now()
	currentDateStr := fmt.Sprintf("_Last Updated: %02d-%02d-%02d %02d:%02d:%02d_", currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(), currentDate.Minute(), currentDate.Second())

	if len(arrival.Services) == 0 {
		result := fmt.Sprintf("There are no more buses coming...Maybe you should hail the cab? ^^a\n %s", currentDateStr)
		return &result, nil
	}

	resultArr := make([]string, 0, len(arrival.Services))
	resultArr = append(resultArr, fmt.Sprintf("*%s, %s*\n```\n", arrival.BusStopCode, stop.Description))

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
			eta, err := nextBusETA(e.NextBus.EstimatedArrival)
			if err != nil {
				return nil, fmt.Errorf("[createBusETAMessage] Problems getting next bus ETA\n%s", err.Error())
			}
			strBus1 = padStart(strconv.Itoa(eta), " ", 3)
		}

		if e.NextBus2.EstimatedArrival == "" {
			strBus2 = "-"
		} else {
			eta, err := nextBusETA(e.NextBus2.EstimatedArrival)
			if err != nil {

				return nil, fmt.Errorf("[createBusETAMessage] Problems getting next bus ETA\n%s", err.Error())
			}
			strBus2 = padStart(strconv.Itoa(eta), " ", 3)
		}
		resultArr = append(resultArr, fmt.Sprintf("%s:%s mins | %s mins\n", strService, strBus1, strBus2))
	}

	resultArr = append(resultArr, "```")
	resultArr = append(resultArr, currentDateStr)

	result := strings.Join(resultArr, "")

	return &result, nil
}

func busStop2Str(busStop int) string {
	return fmt.Sprintf("%05d", busStop)
}

// Returns the number of minutes for the next bus to come
func nextBusETA(estimatedTimeArr string) (int, error) {

	//"EstimatedArrival": "2017-06-05T14:57:09+08:00"
	/*
		let time:string = nextBusthis.EstimatedArrival;
		let t_index:number = time.indexOf("T");
		let plus_index:number = time.indexOf("+");
		time = time.substr(t_index + 1, plus_index);
	*/

	etaDate, err := time.Parse("2006-01-02T15:04:05-07:00", estimatedTimeArr)
	if err != nil {
		return 0, fmt.Errorf("[nextBusETA] Problems parsing time\n%s", err.Error())
	}

	now := time.Now()

	diff := etaDate.Sub(now)
	mins := diff.Minutes()

	if mins > 0 {
		return int(mins), nil
	}

	return 0, nil
}

func getBusSceneInlineRefreshKeyboard(busStop string, timeStamp int) (*TelegramAPI.InlineKeyboardMarkup, error) {

	bytes, err := json.Marshal(BusRefreshCallbackData{
		Cmd:       "refresh",
		BusStop:   busStop,
		TimeStamp: timeStamp,
	})
	if err != nil {
		return nil, fmt.Errorf("[getBusSceneInlineRefreshKeyboard] Problems converting callback data to string\n%s", err.Error())
	}

	result := TelegramAPI.NewInlineKeyboardMarkup(
		TelegramAPI.NewInlineKeyboardRow(
			TelegramAPI.NewInlineKeyboardButtonData("Refresh", string(bytes)),
		),
	)
	return &result, nil

}
