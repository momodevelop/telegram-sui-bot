package scenes

import (
	"fmt"
	"regexp"
	"strconv"
	"telegram-sui-bot/pkg/director"
	"telegram-sui-bot/pkg/repos"
	"time"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneBus struct {
	RepoBusStops *repos.RepoBusStops
}

func NewSceneBus(repoBusStops *repos.RepoBusStops) *SceneBus {
	return &SceneBus{
		RepoBusStops: repoBusStops,
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
			busStop, err := this.RepoBusStops.GetBusStop(padStart(busStopCodeStr, "0", 5))
			if err != nil {
				tilt(bot, message, fmt.Errorf("[SceneBus][Process] Can't get bus stop\n%s", err.Error()))
				return
			}
			if busStop == nil {
				bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "That bus stop does not exist!"))
				return
			}

			busArrival, err := this.RepoBusStops.GetBusStopArrivals(busStopCodeStr)
			if err != nil {
				tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong getting arrivals for bus stop number\n%s", err.Error()))
				return
			}
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
		busStop, err := this.RepoBusStops.GetBusStopArrivalsByNearestLocation(message.Location.Latitude, message.Location.Longitude)
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong getting bus stop by nearest location\n%s", err.Error()))
			return
		}
		if busStop == nil {
			bot.Send(TelegramAPI.NewMessage(message.Chat.ID, "That bus stop does not exist!"))
			return
		}

		busArrival, err := this.RepoBusStops.GetBusStopArrivals(busStop.BusStopCode)
		if err != nil {
			tilt(bot, message, fmt.Errorf("[SceneBus][Process] Something wrong getting bus stop arrivals by location\n%s", err.Error()))
			return
		}

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
