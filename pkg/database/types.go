package database

type Table interface {
	Create() string
	Insert() string
}

type BusStopTable struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float32
	Longitude   float32
}
