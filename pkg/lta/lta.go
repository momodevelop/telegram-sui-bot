package lta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

func (this *API) CallBusArrivalv2(busStop string, busNumber string) *BusArrivalv2 {
	path := "ltaodataservice/BusArrivalv2?BusStopCode=" + busStop

	if busNumber == "" {
		path += "&ServiceNo=" + busNumber
	}

	var ret *BusArrivalv2
	err := this.CallAPI2JSON(path, &ret)
	if err != nil {
		log.Panic(err)
	}

	return ret
}

func (this *API) CallBusStops(skip int) *BusStops {
	path := "ltaodataservice/BusStops"
	if skip >= 0 {
		path += "?$skip=" + strconv.Itoa(skip)
	}

	var ret *BusStops
	err := this.CallAPI2JSON(path, &ret)
	if err != nil {
		log.Panic(err)
	}

	return ret
}

func (this *API) CallAPI2JSON(path string, v interface{}) error {
	resp := this.CallAPI(path)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[CallAPI2JSON] Bad Status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	return json.Unmarshal(body, &v)
}

func (this *API) CallAPI(path string) *http.Response {
	fullpath := "http://datamall2.mytransport.sg/" + path

	//log.Printf("[API][CallAPI] %s\n", fullpath)
	req, err := http.NewRequest("GET", fullpath, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Add("AccountKey", this.Token)

	resp, err := this.client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	return resp
}
