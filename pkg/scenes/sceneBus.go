package scenes

import (
	"fmt"
	"regexp"
	"strconv"
	Director "telegram_go_sui_bot/pkg/director"
	Lta "telegram_go_sui_bot/pkg/lta"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneBus struct {
	busAPI *Lta.API
}

func NewSceneBus(ltaToken string) *SceneBus {
	return &SceneBus{
		busAPI: Lta.New(ltaToken),
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

func (obj *SceneBus) Process(session *Director.Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
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
				message := "Invalid bus stop code! Try again..."
				msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
				bot.Send(msg)
			}
			busStopCodeStr := fmt.Sprintf("%05d", busStopCode)
			// TODO: Should check if the bus stop even exists

			resp := obj.busAPI.CallBusArrivalv2(busStopCodeStr, "")
			if len(resp.Services) == 0 {
				message := "There are no more buses coming...Maybe you should hail the cab? ^^a"
				msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
				bot.Send(msg)
			} else {

			}

			//ctx.reply(result.body, result.object);

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
			TelegramAPI.NewKeyboardButton("Send Location!"),
		),
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/poke"),
		),
		TelegramAPI.NewKeyboardButtonRow(
			TelegramAPI.NewKeyboardButton("/exit"),
		),
	)

	/*return {
		keyboard: [
			[
				{ text: 'Send location!', request_location: true }
			],
			[
				{ text: '/help' }
			],
			[
				{ text: '/exit'}
			]
		],
		one_time_keyboard: false
	};*/
}
