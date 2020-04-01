package scenes

import (
	"fmt"
	"io/ioutil"
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
	//TODO
	resp := obj.busAPI.CallBusArrivalv2("02151", "")
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Print(string(body))
	recovery(bot, update)

	message := obj.getGreeting()
	msg := TelegramAPI.NewMessage(update.Message.Chat.ID, message)
	msg.ReplyMarkup = obj.getKeyboard()
	bot.Send(msg)
}

func (obj *SceneBus) Process(session *Director.Session, bot *TelegramAPI.BotAPI, update *TelegramAPI.Update) {
	msg := update.Message.Text
	if msg == "/exit" {
		session.ChangeScene("Main")
		return
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
