package lta

import (
	"log"
	"net/http"
)

type API struct {
	Token  string
	client *http.Client
}

func New(token string) *API {
	ret := &API{
		Token:  token,
		client: &http.Client{},
	}
	return ret
}

func (obj *API) CallBusArrivalv2(busStop string, busNumber string) *http.Response {
	path := "ltaodataservice/BusArrivalv2?BusStopCode=" + busStop

	if busNumber == "" {
		path += "&ServiceNo=" + busNumber
	}

	return obj.CallAPI(path)
}

func (obj *API) CallAPI(path string) *http.Response {
	fullpath := "http://datamall2.mytransport.sg/" + path
	req, err := http.NewRequest("GET", fullpath, nil)
	errCheck("Something wrong with creating a new request", err)

	req.Header.Add("AccountKey", obj.Token)

	resp, err := obj.client.Do(req)
	errCheck("Something wrong with processing the request", err)

	return resp
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Printf("%s", msg)
		log.Panic(err)
	}
}
