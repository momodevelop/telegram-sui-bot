package lta

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (this *API) CallBusArrivalv2(busStop string, busNumber string) (*BusArrivalv2, error) {
	path := "ltaodataservice/BusArrivalv2?BusStopCode=" + busStop

	if busNumber == "" {
		path += "&ServiceNo=" + busNumber
	}

	var ret *BusArrivalv2
	err := this.CallAPI2JSON(path, &ret)
	if err != nil {
		return nil, fmt.Errorf("[LTA][CallBusArrivalv2] Something went wrong\n%s", err.Error())
	}
	return ret, nil
}

func (this *API) CallBusStops(skip int) (*BusStops, error) {
	path := "ltaodataservice/BusStops"
	if skip >= 0 {
		path += "?$skip=" + strconv.Itoa(skip)
	}

	var ret *BusStops
	err := this.CallAPI2JSON(path, &ret)
	if err != nil {
		return nil, fmt.Errorf("[LTA][CallBusStops] Something went wrong\n%s", err.Error())
	}
	return ret, nil
}

func (this *API) CallAPI2JSON(path string, v interface{}) error {
	resp, err := this.CallAPI(path)
	if err != nil {
		return fmt.Errorf("[LTA][CallAPI2JSON] Error calling API\n%s", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("[LTA][CallAPI2JSON] Bad Status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[LTA][CallAPI2JSON] Error converting response body to []byte\n%s", err.Error())
	}

	return json.Unmarshal(body, &v)
}

func (this *API) CallAPI(path string) (*http.Response, error) {
	fullpath := "http://datamall2.mytransport.sg/" + path

	//log.Printf("[API][CallAPI] %s\n", fullpath)
	req, err := http.NewRequest("GET", fullpath, nil)
	if err != nil {
		return nil, fmt.Errorf("[LTA][CallAPI] Something went wrong getting a new request\n%s", err.Error())
	}

	req.Header.Add("AccountKey", this.Token)

	resp, err := this.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("[LTA][CallAPI] Something went wrong doing the request\n%s", err.Error())
	}
	return resp, nil
}
