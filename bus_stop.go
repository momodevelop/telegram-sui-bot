package main 

import (
    "fmt"
)

type BusStopEntry struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float64
	Longitude   float64
}


var GlobalBusStops map[string]*BusStopEntry

func SyncBusStops() error {
    GlobalBusStops = make(map[string]*BusStopEntry)

	fmt.Println("Retrieving bus stops from API!")
	Skip := 0
	TotalStops := 0
	for {
		Response, CallErr := GlobalLtaApi.CallBusStops(Skip)
		if CallErr != nil {
            return CallErr
		}
		if Response != nil && len(Response.Value) > 0 {
			TotalStops += len(Response.Value)
			Skip += 500
			fmt.Printf("Adding %d stops...\n", TotalStops)
			for _, E := range Response.Value {
                Entry := &BusStopEntry{}
				Entry.BusStopCode = E.BusStopCode
				Entry.Description = E.Description
				Entry.Latitude = E.Latitude
				Entry.Longitude = E.Longitude
				Entry.RoadName = E.RoadName

                GlobalBusStops[Entry.BusStopCode] = Entry 
			}
		} else {
			break
		}
	}

	fmt.Printf("%d bus stop entries inserted!\n", TotalStops)
    return nil
}

