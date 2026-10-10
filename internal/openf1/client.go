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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/car_data?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create car data request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch car data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetch car data: unexpected HTTP status %s", resp.Status)
	}

	var data []CarData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode car data: %w", err)
	}
	return data, nil
}
