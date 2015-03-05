package main

import (
	"fmt"
	"net/url"
	"net/http"
	"log"
	"encoding/json"
)

const NEXTBUS_API_URL string = "http://webservices.nextbus.com/service/publicJSONFeed"

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

type args []struct{ key, value string }

func (r Route) String() string {
	return fmt.Sprintf("\t Tag: %s - Name: %s\n", r.Tag, r.Title)
}

func (s StopDetails) String() string {
	return fmt.Sprintf("Stop Tag: %s\n Stop Title: %s\n", s.Tag, s.Title)
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

func main() {
	agency := "ttc"

	routes, err := getRouteList(agency)
	if err != nil {
		log.Fatalf("ERROR", err.Error())
	}
	fmt.Println("Routes: ", routes)

	stops, err := getRouteStops(agency, "510")
	if err != nil {
		log.Fatalf("ERROR", err.Error())
	}
	fmt.Println("Stops: ", stops)

}
