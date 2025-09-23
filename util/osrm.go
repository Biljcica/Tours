package util

import (
    "encoding/json"
    "fmt"
    "net/http"
	"log"
)

type osrmResponse struct {
    Routes []struct {
        Distance float64 `json:"distance"`
        Duration float64 `json:"duration"`
    } `json:"routes"`
}

func GetDistanceAndDurationFromOSRM(lon1, lat1, lon2, lat2 float64, profile string) (float64, float64, error) {
    url := fmt.Sprintf("http://osrm:5000/route/v1/%s/%.6f,%.6f;%.6f,%.6f?overview=false",
        profile ,lon1, lat1, lon2, lat2)
	log.Printf("Calling OSRM URL: %s", url)
    resp, err := http.Get(url)
    if err != nil {
        return 0, 0, err
    }
    defer resp.Body.Close()

    var osrmResp osrmResponse
    if err := json.NewDecoder(resp.Body).Decode(&osrmResp); err != nil {
        return 0, 0, fmt.Errorf("decode failed: %v", err)
    }
    log.Printf("OSRM response: %+v", osrmResp)

    if len(osrmResp.Routes) == 0 {
        return 0, 0, fmt.Errorf("no route returned")
    }

    distanceKm := osrmResp.Routes[0].Distance / 1000.0       // km
    durationMin := osrmResp.Routes[0].Duration / 60.0        // minute

    return distanceKm, durationMin, nil
}
