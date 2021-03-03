package main

import (
    tg "github.com/go-telegram-bot-api/telegram-bot-api"

    "fmt"
    "strconv"
    "strings"
    "time"
    "encoding/json"

)

// SceneTypes
type SceneType = int
const (
    SceneType_Home = iota
    SceneType_Bus
)

// Returns the number of minutes for the next bus to come
func NextBusETA(EstimatedTimeArrival string) (int, error) {

	//"EstimatedArrival": "2017-06-05T14:57:09+08:00"
	/*
		let time:string = nextBusthis.EstimatedArrival;
		let t_index:number = time.indexOf("T");
		let plus_index:number = time.indexOf("+");
		time = time.substr(t_index + 1, plus_index);
	*/

	ETADate, Err := time.Parse("2006-01-02T15:04:05-07:00", EstimatedTimeArrival)
	if Err != nil {
		return 0, fmt.Errorf("[nextBusETA] Problems parsing time\n%s", Err.Error())
	}

	Now := time.Now()

	Diff := ETADate.Sub(Now)
	Mins := Diff.Minutes()

	if Mins > 0 {
		return int(Mins), nil
	}

	return 0, nil
}

func CreateBusETAMessage(Arrival *BusArrivalv2, Stop *BusStopEntry) (string, error) {
	CurDate := time.Now()
	CurDateStr := fmt.Sprintf("_Last Updated: %02d-%02d-%02d %02d:%02d:%02d_", CurDate.Year(), CurDate.Month(), CurDate.Day(), CurDate.Hour(), CurDate.Minute(), CurDate.Second())


	ResultArr := make([]string, 0, len(Arrival.Services))
    Str := fmt.Sprintf("*%s, %s*\n```\n", Arrival.BusStopCode, Stop.Description)
	
	ResultArr = append(ResultArr, Str)

    if len(Arrival.Services) == 0 {
        ResultArr = append(ResultArr, MsgSceneBusNoBus)
	} else {
        for _, E := range Arrival.Services {
            // format:
            // <bus no>\t:\t<time>\t|\t<time>
            // 131\t:\t1min\t|\t2minw
            // 131	:	1min	|	2min
            StrService := PadEnd(E.ServiceNo, " ", 5)
            var StrBus1 string
            var StrBus2 string
            if E.NextBus.EstimatedArrival == "" {
                StrBus1 = "-"
            } else {
                ETA, ETAErr := NextBusETA(E.NextBus.EstimatedArrival)
                if ETAErr != nil {
                    return "", ETAErr
                }
                StrBus1 = PadStart(strconv.Itoa(ETA), " ", 3)
            }

            if E.NextBus2.EstimatedArrival == "" {
                StrBus2 = "-"
            } else {
                ETA, ETAErr := NextBusETA(E.NextBus2.EstimatedArrival)
                if ETAErr != nil {
                    return "", ETAErr
                }
                StrBus2 = PadStart(strconv.Itoa(ETA), " ", 3)
            }
            Str := fmt.Sprintf("%s:%s mins | %s mins\n", StrService, StrBus1, StrBus2)
            ResultArr = append(ResultArr, Str) 
        }
    }

	ResultArr = append(ResultArr, "```")
	ResultArr = append(ResultArr, CurDateStr)

	Result := strings.Join(ResultArr, "")

	return Result, nil
}

type BusRefreshCallbackData struct {
	Cmd       string `json:"cmd"`
	BusStop   string `json:"busStop"`
	TimeStamp int    `json:"timeStamp"`
}

func SceneBusGetInlineRefreshKeyboard(BusStop string, TimeStamp int) (*tg.InlineKeyboardMarkup, error) {
	Bytes, Err := json.Marshal(BusRefreshCallbackData{
		Cmd:       "refresh",
		BusStop:   BusStop,
		TimeStamp: TimeStamp,
	})
	if Err != nil {
		return nil, Err
	}

	Result := tg.NewInlineKeyboardMarkup(
		tg.NewInlineKeyboardRow(
			tg.NewInlineKeyboardButtonData("Refresh", string(Bytes)),
		),
	)
	return &Result, nil

}

func ProcessBusRefreshCallback(Bot *tg.BotAPI, Query *tg.CallbackQuery) {
	// Check if CallbackQuery is parsable to BusRefreshCallbackData
	var Data BusRefreshCallbackData
	Err := json.Unmarshal([]byte(Query.Data), &Data)
	if Err != nil {
        Panik("%v\n", Err)
		return
	}

	if Data.Cmd == "refresh" {
        BS, BSFound := GlobalBusStops[Data.BusStop]
        if !BSFound {
			Bot.Send(tg.NewEditMessageText(
				Query.Message.Chat.ID,
				Query.Message.MessageID,
				"That bus stop does not exist!",
			))
        }

		A, Err := GlobalLtaApi.CallBusArrival(Data.BusStop, "")
		if Err != nil {
            Panik("%v\n", Err)
		}

        Now := time.Now().Nanosecond() 
		KB, Err := SceneBusGetInlineRefreshKeyboard(Data.BusStop, Now)
		if Err != nil {
            Panik("Something wrong creating refresh keyboard\n")
		}

		Msg := tg.NewEditMessageText(Query.Message.Chat.ID, Query.Message.MessageID, "")
		Msg.ReplyMarkup = KB
		Msg.ParseMode = "markdown"
        Reply, ReplyErr := CreateBusETAMessage(A, BS)
        if ReplyErr != nil {
            Panik("Error creating bus message: %s\n", ReplyErr)
        }
        Msg.Text = Reply
		Msg.ParseMode = "markdown"
		Bot.Send(Msg)

        CallbackConf := tg.NewCallback(Query.ID, "")
        Bot.AnswerCallbackQuery(CallbackConf)
	}


}

func SceneBusProcess(Sess *Session, Bot *tg.BotAPI, Msg *tg.Message) {
    if len(Msg.Text) > 0 {
        switch Msg.Text {
        case "/exit":
            Sess.CurrentScene = SceneType_Home
        case "/update":
            if Err := SyncBusStops(); Err != nil {
                Panik("Cannot sync buses\n")
            } else {
                NewMsg := tg.NewMessage(Msg.Chat.ID, "Updated")
                NewMsg.ReplyMarkup = SceneBusGetKeyboard()
                Bot.Send(NewMsg)
            } 

        default: 
            Matches := GlobalBusNumRegex.FindAllString(Msg.Text, -1)
            if len(Matches) > 0 {
                BusStopCode := Matches[0]
                
                BS, BSFound := GlobalBusStops[BusStopCode]
                if !BSFound {
                    NewMsg := tg.NewMessage(Msg.Chat.ID, MsgSceneBusCannotFindBus)
                    NewMsg.ReplyMarkup = SceneBusGetKeyboard()
                    Bot.Send(NewMsg)
                    return
                }
                A, CallErr := GlobalLtaApi.CallBusArrival(BusStopCode, "")
                if CallErr != nil {
                    Panik("CallBusArrival error: %s\n", CallErr)
                }

                Reply, ReplyErr := CreateBusETAMessage(A, BS)
                if ReplyErr != nil {
                    Panik("Error creating bus message: %s\n", ReplyErr)
                }

                Now := time.Now().Nanosecond()
                KB, KBErr := SceneBusGetInlineRefreshKeyboard(BS.BusStopCode, Now)
                if KBErr != nil {
                    Panik("Something wrong creating refresh keyboard\n")
                }

                NewMsg := tg.NewMessage(Msg.Chat.ID, Reply)
                NewMsg.ParseMode = "markdown"
                NewMsg.ReplyMarkup = KB
                Bot.Send(NewMsg)
                return
            } 

            NewMsg := tg.NewMessage(Msg.Chat.ID, MsgSceneBusGreeting)
            NewMsg.ReplyMarkup = SceneBusGetKeyboard() 
            Bot.Send(NewMsg)
            return
        }
    } else if Msg.Location != nil {
        // Find the bus stop nearest to location
        // TODO: Optimize this?
        var BS *BusStopEntry = nil
        Long := Msg.Location.Longitude
        Lat := Msg.Location.Latitude
        ShortestDist := float64(0)
        for _, E := range GlobalBusStops {
            LongDiff := E.Longitude - Long
            LatDiff := E.Latitude - Lat
            Distance := LongDiff * LongDiff + LatDiff * LatDiff 
            if BS == nil || Distance < ShortestDist {
                ShortestDist = Distance
                BS = E
            }
        }
		if BS == nil {
            NewMsg := tg.NewMessage(Msg.Chat.ID, MsgSceneBusNoNearbyBusStops)
			Bot.Send(NewMsg)
			return
		}
   
        fmt.Printf("%v\n", BS)
		A, Err := GlobalLtaApi.CallBusArrival(BS.BusStopCode, "")
		if Err != nil {
            Panik("Something wrong with CallBusArrival: %v\n", Err)
		}

		Reply, ReplyErr := CreateBusETAMessage(A, BS)
		if ReplyErr != nil {
            Panik("Error creating bus message: %s\n", ReplyErr)
		}

        Now := time.Now().Nanosecond()
		KB, KBErr := SceneBusGetInlineRefreshKeyboard(BS.BusStopCode, Now)
		if KBErr != nil {
            Panik("Something wrong creating refresh keyboard\n")
		}
		NewMsg := tg.NewMessage(Msg.Chat.ID, Reply)
		NewMsg.ParseMode = "markdown"
		NewMsg.ReplyMarkup = KB
		Bot.Send(NewMsg)
    }
}

func SceneBusGetKeyboard() tg.ReplyKeyboardMarkup {
	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.KeyboardButton{
				Text: "Send Location!",
				RequestLocation: true,
			},
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/poke"),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/exit"),
		),
	)
}
func SceneHomeProcess(Sess *Session, Bot *tg.BotAPI, Msg *tg.Message) {
    switch Msg.Text {
        case "/bus":
            Sess.CurrentScene = SceneType_Bus
            SceneBusProcess(Sess, Bot, Msg)
            
        case "/food":
            Reply := fmt.Sprintf(MsgFoodRecommend, RandomMsg(MsgEatResponses))
            NewMsg := tg.NewMessage(Msg.Chat.ID, Reply) 
            NewMsg.ReplyMarkup = SceneHomeGetKeyboard()
            Bot.Send(NewMsg)
        default: 
            NewMsg := tg.NewMessage(Msg.Chat.ID, MsgSceneHomeGreeting)
            NewMsg.ReplyMarkup = SceneHomeGetKeyboard() 
            Bot.Send(NewMsg)
    }
}

func SceneHomeGetKeyboard() tg.ReplyKeyboardMarkup {
	return tg.NewReplyKeyboard(
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/bus"),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/food"),
		),
		tg.NewKeyboardButtonRow(
			tg.NewKeyboardButton("/poke"),
		),
	)
}
