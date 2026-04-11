// Package daemon provides client utilities for communicating with the daemon.
package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client provides methods to interact with the daemon API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new daemon client
func NewClient(port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://localhost:%d", port),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// IsRunning checks if the daemon is running
func (c *Client) IsRunning() bool {
	resp, err := c.httpClient.Get(c.baseURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// doRequest performs an HTTP request and returns the response body
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}

// decodeJSON decodes a JSON response
func decodeJSON(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, errResp["error"])
	}
	// Handle 204 No Content - no body to decode
	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// GetHealth gets daemon health status
func (c *Client) GetHealth() (map[string]string, error) {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	err = decodeJSON(resp, &result)
	return result, err
}

// GetTimers gets all timers
func (c *Client) GetTimers() ([]map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/timers", nil)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// GetRunningTimers gets running timers
func (c *Client) GetRunningTimers() ([]map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/timers/running", nil)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// CreateTimer creates a new timer
func (c *Client) CreateTimer(label string, durationMs int64, start bool) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"label":       label,
		"duration_ms": durationMs,
		"start":       start,
	}
	resp, err := c.doRequest("POST", "/timers", req)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// StartTimer starts a timer
func (c *Client) StartTimer(id int64) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/timers/%d/start", id), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// PauseTimer pauses a timer
func (c *Client) PauseTimer(id int64) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/timers/%d/pause", id), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// GetStopwatches gets all stopwatches
func (c *Client) GetStopwatches() ([]map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/stopwatches", nil)
	if err != nil {
		return nil, err
	}
	var result []map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// CreateStopwatch creates a new stopwatch
func (c *Client) CreateStopwatch(label string, start bool) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"label": label,
		"start": start,
	}
	resp, err := c.doRequest("POST", "/stopwatches", req)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// StartStopwatch starts a stopwatch
func (c *Client) StartStopwatch(id int64) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/stopwatches/%d/start", id), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// StopStopwatch stops a stopwatch
func (c *Client) StopStopwatch(id int64) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", fmt.Sprintf("/stopwatches/%d/stop", id), nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// CreateJournalEntry creates a journal entry
func (c *Client) CreateJournalEntry(log string) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"log": log,
	}
	resp, err := c.doRequest("POST", "/journal", req)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	err = decodeJSON(resp, &result)
	return result, err
}

// GetSettings gets all settings
func (c *Client) GetSettings() (map[string]string, error) {
	resp, err := c.doRequest("GET", "/settings", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	err = decodeJSON(resp, &result)
	return result, err
}

// SetSetting sets a setting value
func (c *Client) SetSetting(key, value string) error {
	req := map[string]interface{}{
		"value": value,
	}
	resp, err := c.doRequest("PUT", fmt.Sprintf("/settings/%s", key), req)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// PlayAmbience starts ambient sound playback
func (c *Client) PlayAmbience() error {
	resp, err := c.doRequest("POST", "/ambient/play", nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// StopAmbience stops ambient sound playback
func (c *Client) StopAmbience() error {
	resp, err := c.doRequest("POST", "/ambient/stop", nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// SetAmbienceVolume sets ambient volume
func (c *Client) SetAmbienceVolume(volume float64) error {
	req := map[string]interface{}{
		"volume": volume,
	}
	resp, err := c.doRequest("PUT", "/ambient/volume", req)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// RefreshAmbience tells the daemon to reload config and refresh playback
func (c *Client) RefreshAmbience() error {
	resp, err := c.doRequest("POST", "/ambient/refresh", nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// StopDaemon sends a shutdown request to the daemon
func (c *Client) StopDaemon() error {
	resp, err := c.doRequest("POST", "/daemon/shutdown", nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// DeleteTimer deletes a timer
func (c *Client) DeleteTimer(id int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/timers/%d", id), nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}

// DeleteStopwatch deletes a stopwatch
func (c *Client) DeleteStopwatch(id int64) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/stopwatches/%d", id), nil)
	if err != nil {
		return err
	}
	return decodeJSON(resp, &struct{}{})
}
