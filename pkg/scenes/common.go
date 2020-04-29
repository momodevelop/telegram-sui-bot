package scenes

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-sui-bot/pkg/lta"
	"telegram-sui-bot/pkg/repos"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

func recovery(bot *TelegramAPI.BotAPI, message *TelegramAPI.Message) {
	if r := recover(); r != nil {
		msg := TelegramAPI.NewMessage(message.Chat.ID, "Oops, something went wrong ><")
		bot.Send(msg)

		log.Println("Stop panicking: ", r)
	}
}

func tilt(bot *TelegramAPI.BotAPI, message *TelegramAPI.Message, err error) {
	log.Printf("%s", err.Error())
	msg := TelegramAPI.NewMessage(message.Chat.ID, "Oops, something went wrong ><")
	bot.Send(msg)
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

// helper functions
func createBusETAMessage(arrival *lta.BusArrivalv2, stop *repos.BusStopTable) (*string, error) {
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
