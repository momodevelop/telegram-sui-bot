package main

import (
    tg "github.com/go-telegram-bot-api/telegram-bot-api"
    
    "net/http"
    "strings"
    "regexp"
	"fmt"
	"io/ioutil"
)

type UserId = int
type Session struct {
    CurrentScene SceneType 
}

var (
    GlobalBusNumRegex *regexp.Regexp
    GlobalSessions map[UserId]*Session
    GlobalHttpClient *http.Client
    GlobalLtaApi *LtaApi
)

func PadStart(Str string, Item string, Count int) string {
	PadAmt := Count - len(Str)
	if PadAmt > 0 {
		return strings.Repeat(Item, PadAmt) + Str
	}
	return Str
}

func PadEnd(Str string, Item string, Count int) string {
	PadAmount := Count - len(Str)
	if PadAmount > 0 {
		return Str + strings.Repeat(Item, PadAmount)
	}
	return Str
}

func Panik(Format string, a ...interface{}) {
	panic(fmt.Sprintf(Format, a...))
}

func Kalm(Bot *tg.BotAPI, Msg *tg.Message) {
	if R := recover(); R != nil {
		fmt.Printf("Recovered: %v\n", R)
        NewMsg := tg.NewMessage(Msg.Chat.ID, MsgGenericFail) 
        Bot.Send(NewMsg)
	}
}


func ProcessCallback(Bot *tg.BotAPI, Query *tg.CallbackQuery) {
    // There is only one callback for now.
    go func() {
        Kalm(Bot, Query.Message)
        ProcessBusRefreshCallback(Bot, Query)
    }()
}

func ProcessMessage(Bot *tg.BotAPI, Msg *tg.Message) {
	User := Msg.From

	// Check if session exists. If it does not, create new session
	Sess, IsOldUser := GlobalSessions[User.ID]
	if !IsOldUser {
		var ok bool
		GlobalSessions[User.ID] = &Session{}
		Sess, ok = GlobalSessions[User.ID]
		if !ok {
			Panik("Cannot create session for %d", User.ID)
		}
	}

    // Redirect based on session
    switch Sess.CurrentScene {
        case SceneType_Home:
            go func() {
                Kalm(Bot, Msg)
                SceneHomeProcess(Sess, Bot, Msg) 
            }()
        case SceneType_Bus: 
            go func() {
                Kalm(Bot, Msg)
                SceneBusProcess(Sess, Bot, Msg) 
            }()
        default:
            Panik("Unknown scene! %d\n", Sess.CurrentScene)
        }    
}

func main() {
    GlobalSessions = make(map[UserId]*Session)
	TelegramToken, TelegramTokenErr := ioutil.ReadFile("TOKEN")
	if TelegramTokenErr != nil {
		Panik("Cannot read or find TOKEN file\n")
	}

    var Err error
    GlobalBusNumRegex, Err = regexp.Compile(`\d+`)
    if Err != nil {
        Panik("Cannot compile BusNumRegex")
    }

    LtaToken, LtaTokenErr := ioutil.ReadFile("LTA_TOKEN")
	if LtaTokenErr != nil {
		Panik("Cannot read or find LTA_TOKEN file\n")
	}

    // Lta API init
    GlobalLtaApi = NewLtaApi(string(LtaToken))

    // Get bus stops
    if SyncErr := SyncBusStops(); SyncErr != nil {
        Panik("Cannot get bus stop\n")
    }

    // Telegram Bot initialization
    Bot, BotErr := tg.NewBotAPI(string(TelegramToken))
    if BotErr != nil {
        Panik("%v", BotErr)
    }
    fmt.Printf("Authorized on account %s", Bot.Self.UserName)
    U := tg.NewUpdate(0)
    U.Timeout = 60
    Updates, UpdatesErr := Bot.GetUpdatesChan(U)
    if UpdatesErr != nil {
        Panik("%v", UpdatesErr)
    }

    for Update := range Updates {
        if Update.CallbackQuery != nil {
            ProcessCallback(Bot, Update.CallbackQuery)
        } else if Update.Message != nil {
            ProcessMessage(Bot, Update.Message)
        }
    }
}
