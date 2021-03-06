package main

import (
	"fmt"
	"net/url"
	"net/http"
	"log"
	"encoding/json"
	"time"
	"flag"
)

const NEXTBUS_API_URL string = "http://webservices.nextbus.com/service/publicJSONFeed"

type Agency struct {
	Title string
	Tag string
	RegionTitle string
	ShortTitle string
}

type Route struct {
	Tag string
	Title string
}

type RouteDetails struct {
	Title string
	Tag string
	LatMin string
	LonMin string
	LatMax string
	LonMax string
	Stop []StopDetails
	Direction []DirectionDetails
	Path []struct {
		Point []struct {
			Lat string
			Lon string
		}
	}
}

type DirectionDetails struct {
	Title string
	Tag string
	Name string
	Branch string
	Stop []struct {
		Tag string
	}
}

type StopDetails struct {
	Title string
	StopId string
	Tag string
	Lat string
	Lon string
}

type Predictions struct {
	AgencyTitle string
	RouteTag string
	RouteTitle string
	StopTitle string
	StopTag string
	Direction struct {
		Title string
		Prediction []PredictionDetails
	}
}

type PredictionDetails struct {
	IsDeparture string
	Minutes string
	Seconds string
	TripTag string
	Vehicle string
	Block string
	Branch string
	DirTag string
	EpochTime string
}

type Schedule struct {
	Title string
	Tag string
	Direction string
	ServiceClass string
	ScheduleClass string

	Header struct {
		Stop []struct {
			Content string
			Tag string
		}
	}
	Tr []struct {
		BlockId string
		Stop []struct {
			Content string
			Tag string
			EpochTime string
		}
	}
}

type VehicleLocations struct {
	LastTime struct {
		Time string
	}
	Vehicle []VehicleLocation
}

type VehicleLocation struct {
		Id string
		RouteTag string
		DirTag string
		Predictable string
		Lon string
		Lat string
		Heading string
		SecsSinceReport string
}

type args []struct {
	key string
	value string
}

func (r Route) String() string {
	return fmt.Sprintf("\n\t Title: %s - Tag: %s", r.Title, r.Tag)
}

func (a Agency) String() string {
	return fmt.Sprintf("\n\t Agency: %s - Tag: %s\n\t Region: %s Short: %s",
		a.Title, a.Tag, a.RegionTitle, a.ShortTitle)
}

func (r RouteDetails) String() string {
	return fmt.Sprintf("\n\t Title: %s\n\t Tag: %s\n\t Stops:\n\t %s\n\t Directions:\n\t %s",
		r.Title, r.Tag, r.Stop, r.Direction)
}

func (d DirectionDetails) String() string {
	return fmt.Sprintf("\n\t Title: %s - Tag: %s", d.Title, d.Tag)
}

func (s StopDetails) String() string {
	return fmt.Sprintf("\n\t Title: %s - Tag: %s", s.Title, s.Tag)
}

func (p Predictions) String() string {
	return fmt.Sprintf("\n\t Route: %s - Tag: %s\n\t Stop: %s - Tag: %s\n\tDirections:\n\t %s",
		p.RouteTitle, p.RouteTag, p.StopTitle, p.StopTag, p.Direction)
}

func (p PredictionDetails) String() string {
	return fmt.Sprintf("\n\t Vehicle: %s - Block: %s - Branch: %s - Direction: %s\n\t Minutes: %s - Seconds: %s",
		p.Vehicle, p.Block, p.Branch, p.DirTag, p.Minutes, p.Seconds)
}

func (v VehicleLocations) String() string {
	return fmt.Sprintf("\n\t Vehicles: %s\n\t Last Update: %s", v.Vehicle, v.LastTime)
}

func (v VehicleLocation) String() string {
	return fmt.Sprintf("\n\t Vehicle ID: %s - Direction Tag: %s\n\t Route Tag: %s - Seconds Since: %s\n\t Lon: %s - Lat: %s - Heading: %s",
		v.Id, v.DirTag, v.RouteTag, v.SecsSinceReport, v.Lon, v.Lat, v.Heading)
}

func (a args) makeUrl(command string) string {
	apiUrl, err := url.Parse(NEXTBUS_API_URL)
	if err != nil {
		log.Fatalf("API URL is not valid.", err.Error())
	}
	parameters := url.Values{}
	parameters.Add("command", command)
	for _, arg := range a {
		parameters.Add(arg.key, arg.value)
	}
	apiUrl.RawQuery = parameters.Encode()
	return apiUrl.String()
}

func fetchData(url string, d interface{}) error {
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("HTTP request failed.", err.Error())
		return err
	}
	err = json.NewDecoder(resp.Body).Decode(d)

	if err != nil {
		log.Fatalf("JSON decoding failed.", err.Error())
		return err
	}
	return nil
}

func getAgencyList() ([]Agency, error) {
	var data struct{ Agency []Agency }
	err := fetchData(args{}.makeUrl("agencyList"), &data)
	return data.Agency, err
}

func getRouteList(agency string) ([]Route, error) {
	args := args{{"a", agency}}
	var data struct{ Route []Route }
	err := fetchData(args.makeUrl("routeList"), &data)
	return data.Route, err
}

func getRouteStops(agency, route string) (RouteDetails, error) {
	args := args{{"a", agency}, {"r", route}}
	var data struct{ Route RouteDetails }
	err := fetchData(args.makeUrl("routeConfig"), &data)
	return data.Route, err
}

func getPredictions(agency, route, stopTag string) (Predictions, error) {
	args := args{{"a", agency}, {"r", route}, {"s", stopTag}}
	var data struct{ Predictions Predictions }
	err := fetchData(args.makeUrl("predictions"), &data)
	return data.Predictions, err
}

func getSchedule(agency, route string) ([]Schedule, error) {
	args := args{{"a", agency}, {"r", route}}
	var data struct{ Route []Schedule }
	err := fetchData(args.makeUrl("schedule"), &data)
	return data.Route, err
}

func getVehicleLocations(agency, route string) (VehicleLocations, error) {
	args := args{{"a", agency}, {"r", route}, {"t", fmt.Sprint(time.Now().Unix())}}
	var data VehicleLocations
	err := fetchData(args.makeUrl("vehicleLocations"), &data)
	return data, err
}

func main() {
	method := flag.String("method", "agencies", "Available methods: agencies, routes, stops, locations, predictions, schedule")
	agencyCode := flag.String("agency", "ttc", "Toronto Transit Commission")
	routeCode := flag.String("route", "510", "510 Default")
	stopID := flag.String("stop", "14339", "Stop ID")

	flag.Parse();

	var err error
	var display string
	var data interface{}

	switch *method {
	case "locations":
		data, err = getVehicleLocations(*agencyCode, *routeCode)
		display = "Vehicle Locations: "
	case "routes":
		data, err = getRouteList(*agencyCode)
		display = "Route List: "
	case "stops":
		data, err = getRouteStops(*agencyCode, *routeCode)
		display = "Route Stops: "
	case "predictions":
		data, err = getPredictions(*agencyCode, *routeCode, *stopID)
		display = "Predictions: "
	case "schedule":
		data, err = getSchedule(*agencyCode, *routeCode)
		display = "Schedule: "
	case "agencies":
		fallthrough
	default:
		data, err = getAgencyList()
		display = "Agency List: "
	}

	// locations, err := getVehicleLocations(*agencyCode, *routeCode)
	if err != nil {
		log.Fatalf("ERROR", err.Error())
	}

	fmt.Println(display, data)
}
