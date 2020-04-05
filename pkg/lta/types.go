package lta

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
