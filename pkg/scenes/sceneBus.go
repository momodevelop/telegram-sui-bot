package scenes

import (
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

func (obj *SceneBus) Name() string {
	return "Bus"
}

func (obj *SceneBus) Greet(bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	message := obj.getGreeting()
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = obj.getKeyboard()
	bot.Send(msg)
}

func (obj *SceneBus) Process(session *director.Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
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
				sendSimpleReply("Invalid bus stop code! Try again...", bot, update)
				return
			}
			busStopCodeStr := obj.busStop2Str(busStopCode)
			busStop := obj.db.GetBusStop(padStart(busStopCodeStr, "0", 5))
			if busStop == nil {
				sendSimpleReply("That bus stop does not exist!", bot, update)
				return
			}

			busArrival := obj.busAPI.CallBusArrivalv2(busStopCodeStr, "")
			if len(busArrival.Services) == 0 {
				sendSimpleReply("There are no more buses coming...Maybe you should hail the cab? ^^a", bot, update)
				return
			}

			message := obj.createBusETAMessage(busArrival, busStop)
			msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
			msg.ParseMode = "markdown"
			bot.Send(msg)

			return
		}

		obj.Greet(bot, update)

	}

}

func (obj *SceneBus) getGreeting() string {
	return "I can help you check for the buses at your bus stop...!\nJust key in the bus stop number or send me your location and I'll try to find the timings ASAP!（｀・ω・´）\n"
}

func (obj *SceneBus) getKeyboard() TelegramAPI.ReplyKeyboardMarkup {
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
func (obj *SceneBus) createBusETAMessage(arrival *lta.BusArrivalv2, stop *database.BusStopTable) string {
	currentDate := time.Now()
	currentDateStr := fmt.Sprintf("_Last Updated: %d-%d-%d %d:%d:%d_", currentDate.Year(), currentDate.Month(), currentDate.Day(), currentDate.Hour(), currentDate.Minute(), currentDate.Second())

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
			strBus1 = padStart(strconv.Itoa(obj.nextBusETA(e.NextBus.EstimatedArrival)), " ", 3)
		}

		if e.NextBus2.EstimatedArrival == "" {
			strBus2 = "-"
		} else {
			strBus2 = padStart(strconv.Itoa(obj.nextBusETA(e.NextBus2.EstimatedArrival)), " ", 3)
		}
		resultArr = append(resultArr, fmt.Sprintf("%s:%s mins | %s mins\n", strService, strBus1, strBus2))
	}

	resultArr = append(resultArr, "```")
	resultArr = append(resultArr, currentDateStr)

	return strings.Join(resultArr, "")
}

func (obj *SceneBus) busStop2Str(busStop int) string {
	return fmt.Sprintf("%05d", busStop)
}

// Returns the number of minutes for the next bus to come
func (obj *SceneBus) nextBusETA(estimatedTimeArr string) int {

	//"EstimatedArrival": "2017-06-05T14:57:09+08:00"
	/*
		let time:string = nextBusObj.EstimatedArrival;
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
