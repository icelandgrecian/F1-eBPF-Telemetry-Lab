package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "https://api.openf1.org/v1"

// Client retrieves historical telemetry from OpenF1.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a client with a timeout for each HTTP request.
func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// CarData returns telemetry for a session and driver.
func (c *Client) CarData(ctx context.Context, sessionKey int, driverNumber int) ([]CarData, error) {
	query := url.Values{}
	query.Set("session_key", strconv.Itoa(sessionKey))
	query.Set("driver_number", strconv.Itoa(driverNumber))

	var data []CarData
	if err := c.get(ctx, "/car_data", query, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// Sessions lists sessions filtered by year, or by a specific session key.
// A zero filter is omitted from the request.
func (c *Client) Sessions(ctx context.Context, year, sessionKey int) ([]Session, error) {
	query := url.Values{}
	if year > 0 {
		query.Set("year", strconv.Itoa(year))
	}
	if sessionKey > 0 {
		query.Set("session_key", strconv.Itoa(sessionKey))
	}
	var data []Session
	if err := c.get(ctx, "/sessions", query, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// Drivers lists the drivers participating in the selected session.
func (c *Client) Drivers(ctx context.Context, sessionKey int) ([]Driver, error) {
	query := url.Values{"session_key": {strconv.Itoa(sessionKey)}}
	var data []Driver
	if err := c.get(ctx, "/drivers", query, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Client) get(ctx context.Context, path string, query url.Values, destination any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path+"?"+query.Encode(), nil)
	if err != nil {
		return fmt.Errorf("create %s request: %w", path, err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("fetch %s: unexpected HTTP status %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}
