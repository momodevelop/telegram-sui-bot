package lta

import (
	"encoding/json"
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
	errCheck("[CallBusArrivalv2] Something went wrong", err)

	return ret
}

func (this *API) CallBusStops(skip int) *BusStops {
	path := "/ltaodataservice/BusStops"
	if skip >= 0 {
		path += "?$skip=" + strconv.Itoa(skip)
	}

	var ret *BusStops
	err := this.CallAPI2JSON(path, &ret)
	errCheck("[CallBusArrivalv2] Something went wrong", err)

	return ret
}

func (this *API) CallAPI2JSON(path string, v interface{}) error {
	resp := this.CallAPI(path)
	body, err := ioutil.ReadAll(resp.Body)
	errCheck("[CallAPI2JSON] Error converting response body to []byte", err)

	return json.Unmarshal(body, &v)
}

func (this *API) CallAPI(path string) *http.Response {
	fullpath := "http://datamall2.mytransport.sg/" + path

	//log.Printf("[API][CallAPI] %s\n", fullpath)
	req, err := http.NewRequest("GET", fullpath, nil)
	errCheck("Something wrong with creating a new request", err)

	req.Header.Add("AccountKey", this.Token)

	resp, err := this.client.Do(req)
	errCheck("Something wrong with processing the request", err)

	return resp
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Printf("%s", msg)
		log.Panic(err)
	}
}
