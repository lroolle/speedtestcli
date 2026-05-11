package cfbackend

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testMetaJSON = `{
	"clientIp": "203.0.113.42",
	"asn": 13335,
	"asOrganization": "Cloudflare Inc",
	"country": "US",
	"city": "San Francisco",
	"region": "California",
	"postalCode": "94107",
	"latitude": "37.7749",
	"longitude": "-122.4194",
	"colo": {
		"iata": "SFO",
		"lat": 37.6213,
		"lon": -122.379,
		"cca2": "US",
		"region": "North America",
		"city": "San Francisco"
	}
}`

func TestFetchMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/meta" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testMetaJSON))
	}))
	defer srv.Close()

	info, err := fetchMeta(context.Background(), srv.Client(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.ClientIP != "203.0.113.42" {
		t.Errorf("client_ip: expected 203.0.113.42, got %s", info.ClientIP)
	}
	if info.ASN != 13335 {
		t.Errorf("asn: expected 13335, got %d", info.ASN)
	}
	if info.ASOrg != "Cloudflare Inc" {
		t.Errorf("as_org: expected Cloudflare Inc, got %s", info.ASOrg)
	}
	if info.City != "San Francisco" {
		t.Errorf("city: expected San Francisco, got %s", info.City)
	}
	if info.Colo.IATA != "SFO" {
		t.Errorf("colo.iata: expected SFO, got %s", info.Colo.IATA)
	}
	if info.Latitude == 0 {
		t.Error("latitude should be parsed from string")
	}
}

func TestFetchMeta_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	_, err := fetchMeta(context.Background(), srv.Client(), srv.URL)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}
