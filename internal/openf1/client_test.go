package openf1

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestMetadataRequests(t *testing.T) {
	var requests []string
	client := &Client{httpClient: &http.Client{Transport: transportFunc(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.RequestURI())
		body := `[{"session_key":42,"meeting_key":8,"session_name":"Race","circuit_short_name":"Test","year":2024}]`
		if req.URL.Path == "/v1/drivers" {
			body = `[{"driver_number":7,"full_name":"API Driver","team_name":"API Team"}]`
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	sessions, err := client.Sessions(context.Background(), 2024, 0)
	if err != nil || len(sessions) != 1 || sessions[0].SessionName != "Race" {
		t.Fatalf("sessions=%+v err=%v", sessions, err)
	}
	_, err = client.Sessions(context.Background(), 0, 42)
	if err != nil {
		t.Fatal(err)
	}
	drivers, err := client.Drivers(context.Background(), 42)
	if err != nil || len(drivers) != 1 || drivers[0].FullName != "API Driver" {
		t.Fatalf("drivers=%+v err=%v", drivers, err)
	}
	want := []string{"/v1/sessions?year=2024", "/v1/sessions?session_key=42", "/v1/drivers?session_key=42"}
	for i := range want {
		if requests[i] != want[i] {
			t.Fatalf("request %d=%q, want %q", i, requests[i], want[i])
		}
	}
}
