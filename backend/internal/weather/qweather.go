package weather

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// QWeatherClient handles communication with QWeather API
type QWeatherClient struct {
	apiHost string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// NewQWeatherClient creates a new QWeather API client
func NewQWeatherClient(apiHost, apiKey string, timeout time.Duration) *QWeatherClient {
	return &QWeatherClient{
		apiHost: apiHost,
		apiKey:  apiKey,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// QWeatherResponse represents the API response structure
type QWeatherResponse struct {
	Code       string       `json:"code"`
	UpdateTime string       `json:"updateTime"`
	FxLink     string       `json:"fxLink"`
	Daily      []DailyData  `json:"daily"`
	Refer      ReferData    `json:"refer"`
}

// DailyData represents daily forecast data
type DailyData struct {
	FxDate    string `json:"fxDate"`
	TempMax   string `json:"tempMax"`
	TempMin   string `json:"tempMin"`
	Precip    string `json:"precip"`
	Humidity  string `json:"humidity"`
	Pressure  string `json:"pressure"`
	TextDay   string `json:"textDay"`
	TextNight string `json:"textNight"`
}

// ReferData represents data source information
type ReferData struct {
	Sources []string `json:"sources"`
	License []string `json:"license"`
}

// GetDailyForecast retrieves N-day weather forecast
// days: 3, 7, 10, 15, or 30
// latitude, longitude: location coordinates
func (c *QWeatherClient) GetDailyForecast(days int, latitude, longitude float64) (*QWeatherResponse, error) {
	// Validate days parameter
	validDays := map[int]bool{3: true, 7: true, 10: true, 15: true, 30: true}
	if !validDays[days] {
		return nil, fmt.Errorf("invalid days parameter: %d, must be 3, 7, 10, 15, or 30", days)
	}

	// Build URL with key parameter
	location := fmt.Sprintf("%.2f,%.2f", longitude, latitude)
	url := fmt.Sprintf("https://%s/v7/weather/%dd?location=%s&key=%s", c.apiHost, days, location, c.apiKey)

	// Create request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set accept encoding header for gzip compression
	req.Header.Set("Accept-Encoding", "gzip")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Handle gzip encoding
	var reader io.Reader = resp.Body
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	// Parse response
	var weatherResp QWeatherResponse
	if err := json.NewDecoder(reader).Decode(&weatherResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check API response code
	if weatherResp.Code != "200" {
		return nil, fmt.Errorf("API returned error code: %s", weatherResp.Code)
	}

	return &weatherResp, nil
}

// Get15DayForecast is a convenience method to get 15-day forecast
func (c *QWeatherClient) Get15DayForecast(latitude, longitude float64) (*QWeatherResponse, error) {
	return c.GetDailyForecast(15, latitude, longitude)
}
