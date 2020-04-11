package database

type Table interface {
	Create() string
	Insert() string
}

type BusStopTable struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float64
	Longitude   float64
}
