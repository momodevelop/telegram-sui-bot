package main

import ( 
    "math/rand"
)

const MsgSceneHomeGreeting = "Erm...Hi? I'm Tachibana Sui, your humble...er...bot.\nI can help you do a few things, just give me one of the commands:\n---\n/bus to get bus ETA\n/food if you want me to help you decide what to eat"
const MsgGenericFail = "Sorry, something went wrong...contact Momo? (´・ω・`)"


var MsgEatResponses = []string{
	"something with curry",
	"something soupy",
	"something with rice",
	"something with bread",
	"something with noodles",

	// specific
	"udon",
	"soba",
	"sushi",
	"ramen",
	"pasta",
	"pizza",
	"burger",
	"wrap",
	"sandwich",

	// meat
	"something with beef",
	"something with chicken",
	"something with pork",
	"something with fish",
	"something with meat",
	"something with vegetables",
	"something with tofu",

	//cultural
	"Indian",
	"Western",
	"Japanese",
	"Korean",
	"Chinese",
	"Italian",
	"Mexican",
	"Turkish",
	"Local",
}
const MsgFoodRecommend = "You should have **%s**!"
const MsgInvalidBusCode = "Invalid bus stop code! It must have no more than 5 digits! Try again..."
const MsgSceneBusGreeting = "I can help you check for the buses at your bus stop...!\nJust key in the bus stop number or send me your location and I'll try to find the timings ASAP!（｀・ω・´）\n"
const MsgSceneBusCannotFindBus = "That bus stop does not exist!\n"
const MsgSceneBusNoBus = "There are no more buses coming...Maybe you should hail the cab? ^^a\n%s"
const MsgSceneBusNoNearbyBusStops = "There are no nearby bus stops!"

func RandomMsg(MsgArr []string) string {
	RandIndex := rand.Intn(len(MsgArr))
	return MsgArr[RandIndex]
}
