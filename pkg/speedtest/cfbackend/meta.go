package cfbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/lroolle/speedtestcli/pkg/speedtest"
)

type metaResponse struct {
	ClientIP     string `json:"clientIp"`
	ASN          uint64 `json:"asn"`
	ASOrg        string `json:"asOrganization"`
	Country      string `json:"country"`
	City         string `json:"city"`
	Region       string `json:"region"`
	PostalCode   string `json:"postalCode"`
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Colo         coloResponse `json:"colo"`
}

type coloResponse struct {
	IATA    string  `json:"iata"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Country string  `json:"cca2"`
	Region  string  `json:"region"`
	City    string  `json:"city"`
}

func fetchMeta(ctx context.Context, client HTTPDoer, baseURL string) (*speedtest.ConnectionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/meta", nil)
	if err != nil {
		return nil, fmt.Errorf("creating meta request: %w", err)
	}
	req.Header.Set("Referer", baseURL+"/")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching meta: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("meta returned status %d", resp.StatusCode)
	}

	var m metaResponse
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("decoding meta: %w", err)
	}

	lat, _ := strconv.ParseFloat(m.Latitude, 64)
	lon, _ := strconv.ParseFloat(m.Longitude, 64)

	return &speedtest.ConnectionInfo{
		ClientIP:  m.ClientIP,
		ASN:       m.ASN,
		ASOrg:     m.ASOrg,
		Country:   m.Country,
		Region:    m.Region,
		City:      m.City,
		Latitude:  lat,
		Longitude: lon,
		Colo: speedtest.ColoInfo{
			IATA:    m.Colo.IATA,
			City:    m.Colo.City,
			Country: m.Colo.Country,
		},
	}, nil
}
