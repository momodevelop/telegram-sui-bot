package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

type BusStopsValue struct {
	BusStopCode string  `json:"BusStopCode"`
	RoadName    string  `json:"RoadName"`
	Description string  `json:"Description"`
	Latitude    float64 `json:"Latitude"`
	Longitude   float64 `json:"Longitude"`
}

type BusStops struct {
	Value []BusStopsValue `json:"value"`
}

type BusArrivalv2Bus struct {
	OriginCode       string `json:"OriginCode"`
	DestinationCode  string `json:"DestinationCode"`
	EstimatedArrival string `json:"EstimatedArrival"`
	Latitude         string `json:"Latitude"`
	Longitude        string `json:"Longitude"`
	VisitNumber      string `json:"VisitNumber"`
	Load             string `json:"Load"`
	Feature          string `json:"Feature"`
	Type             string `json:"Type"`
}

type BusArrivalv2Services struct {
	ServiceNo string          `json:"ServiceNo"`
	Operator  string          `json:"Operator"`
	NextBus   BusArrivalv2Bus `json:"NextBus"`
	NextBus2  BusArrivalv2Bus `json:"NextBus2"`
	NextBus3  BusArrivalv2Bus `json:"NextBus3"`
}

type BusArrivalv2 struct {
	BusStopCode string                 `json:"BusStopCode"`
	Services    []BusArrivalv2Services `json:"Services"`
}

type LtaApi struct {
	Token  string
	Client *http.Client
}

func NewLtaApi(Token string) *LtaApi {
    return &LtaApi {
        Token: Token,
        Client: &http.Client{},
    }
}

func (L *LtaApi) CallBusArrival(BusStop string, BusNumber string) (Ret *BusArrivalv2, Err error) {
	Path := "ltaodataservice/BusArrivalv2?BusStopCode=" + BusStop

	if BusNumber == "" {
		Path += "&ServiceNo=" + BusNumber
	}

	CallErr := L.Call(Path, &Ret)
	if CallErr != nil {
        Err = CallErr
        return
	}
	return 
}

func (L *LtaApi) CallBusStops(Skip int) (Ret *BusStops, Err error) {
	Path := "ltaodataservice/BusStops"
	if Skip >= 0 {
		Path += "?$skip=" + strconv.Itoa(Skip)
	}

	CallErr := L.Call(Path, &Ret)
	if CallErr != nil {
        Err = CallErr
		return	
    }
	return
}

func (L *LtaApi) Call(Path string, v interface{}) (Err error) {
	Resp, CallErr := L.CallAPI(Path)
	if CallErr != nil {
        Err = CallErr
        return
	}

	if Resp.StatusCode != http.StatusOK {
        Err = fmt.Errorf("Bad Status: %d", Resp.StatusCode)
        return
	}

	Body, ReadBodyErr := ioutil.ReadAll(Resp.Body)
	if ReadBodyErr != nil {
        Err = ReadBodyErr
        return 	
    }

    Err = json.Unmarshal(Body, &v)
    return
}

func (L *LtaApi) CallAPI(Path string) (Resp *http.Response, Err error) {
	FullPath := "http://datamall2.mytransport.sg/" + Path

	//log.Printf("[API][CallAPI] %s\n", fullpath)
	Req, ReqErr := http.NewRequest("GET", FullPath, nil)
	if ReqErr != nil {
	    Err = ReqErr
        return
	}

	Req.Header.Add("AccountKey", L.Token)

    Response, ResErr := L.Client.Do(Req)
	if ResErr != nil {
        Err = ResErr
        return
	}
    Resp = Response
	return
}
