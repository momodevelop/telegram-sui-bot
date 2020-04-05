package lta

import (
	"encoding/json"
	"io/ioutil"
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

func (obj *API) CallBusArrivalv2(busStop string, busNumber string) *BusArrivalv2 {
	path := "ltaodataservice/BusArrivalv2?BusStopCode=" + busStop

	if busNumber == "" {
		path += "&ServiceNo=" + busNumber
	}

	resp := obj.CallAPI(path)
	body, err := ioutil.ReadAll(resp.Body)
	errCheck("[CallBusArrivalv2] Converting response body to []byte", err)

	var ret *BusArrivalv2
	err = json.Unmarshal(body, &ret)
	errCheck("[CallBusArrivalv2] Something while unmarshalling json", err)
	return ret
}

func (this *API) callBusStops(skip int16) {
	/*path := "/ltaodataservice/BusStops"
	if skip >= 0 {
		path += "?$skip=" + skip
	}

	resp := this.callAPI(path)
	resp.Body*/

}

func (this *API) CallAPI(path string) *http.Response {
	fullpath := "http://datamall2.mytransport.sg/" + path
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
