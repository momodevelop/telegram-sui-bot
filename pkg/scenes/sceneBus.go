package scenes

import (
	Director "telegram_go_sui_bot/pkg/director"
	LtaAPI "telegram_go_sui_bot/pkg/lta"

	TelegramAPI "github.com/go-telegram-bot-api/telegram-bot-api"
)

type SceneBus struct {
	busAPI *LtaAPI.API
}

func NewSceneBus(landTransportDataMallToken string) *SceneBus {
	return &SceneBus{
		busAPI: LtaAPI.New(landTransportDataMallToken),
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
	switch update.Message.Text {
	case "/exit":
		session.ChangeScene("Main")
		return
	default:
		obj.Greet(bot, update)
	}

}

func (obj *SceneBus) getGreeting() string {
	return "I can help you check for the buses at your bus stop...!\nJust key in the bus stop number or send me your location and I'll try to find the timings ASAP!（｀・ω・´）\n"
}

func (obj *SceneBus) getKeyboard() TelegramAPI.ReplyKeyboardMarkup {
	return TelegramAPI.NewReplyKeyboard(
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
