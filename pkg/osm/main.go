package osm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Element struct {
	ID  int     `json:"id" msgpack:"id"`
	Lat float64 `json:"lat" msgpack:"lat"`
	Lng float64 `json:"lon" msgpack:"lon"`
}

type Response struct {
	Elements []Element `json:"elements" msgpack:"elements"`
}

// FetchElements searches the Overpass API for water fountains within the provided latitude and longitude
func FetchElements(startLng, startLat, endLng, endLat float64) (map[string]interface{}, error) {
	baseURL := "https://overpass.kumi.systems/api/interpreter/?data="
	bounds := fmt.Sprintf("(%f,%f,%f,%f)", startLat, startLng, endLat, endLng)
	query := fmt.Sprintf(`[out:json][timeout:25];(node["amenity"="drinking_water"]%s;);out body;>;out skel qt;`, bounds)
	u := baseURL + url.QueryEscape(query)
	r, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	data := Response{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&data)
	if err != nil {
		return nil, err
	}
	r.Body.Close()

	result := map[string]interface{}{}
	result["elements"] = data.Elements
	return result, nil
}

type GeoCodeResponse struct {
	BoundingBox []string `json:"boundingbox" msgpack:"boundingbox"`
	Lat         string   `json:"lat" msgpack:"lat"`
	Lon         string   `json:"lon" msgpack:"lon"`
	Name        string   `json:"display_name" msgpack:"display_name"`
}
