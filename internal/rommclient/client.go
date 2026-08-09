package rommclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client handles communication with RomM's REST API.
type Client struct {
	baseURL    string
	user       string
	pass       string
	token      string
	httpClient *http.Client
}

// HeartbeatResponse maps RomM /api/heartbeat response payload.
type HeartbeatResponse struct {
	System struct {
		Version string `json:"VERSION"`
	} `json:"SYSTEM"`
	MetadataSources map[string]interface{} `json:"METADATA_SOURCES"`
}

// PlatformResponse maps platform metrics and info fields from /api/platforms.
type PlatformResponse struct {
	ID            interface{} `json:"id"`
	Slug          string      `json:"slug"`
	Name          string      `json:"name"`
	DisplayName   string      `json:"display_name"`
	Category      string      `json:"category"`
	Generation    interface{} `json:"generation"`
	FamilyName    string      `json:"family_name"`
	FamilySlug    string      `json:"family_slug"`
	IgdbID        interface{} `json:"igdb_id"`
	MobyID        interface{} `json:"moby_id"`
	SsID          interface{} `json:"ss_id"`
	RaID          interface{} `json:"ra_id"`
	HasheousID    interface{} `json:"hasheous_id"`
	TgdbID        interface{} `json:"tgdb_id"`
	LaunchboxID   interface{} `json:"launchbox_id"`
	IsIdentified  interface{} `json:"is_identified"`
	MissingFromFS interface{} `json:"missing_from_fs"`
	RomCount      interface{} `json:"rom_count"`
	FSSizeBytes   interface{} `json:"fs_size_bytes"`
	FirmwareCount interface{} `json:"firmware_count"`
}

// NewClient creates a new RomM API client.
func NewClient(baseURL, user, pass, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		user:    user,
		pass:    pass,
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// HasURL checks if a target RomM base URL is configured.
func (c *Client) HasURL() bool {
	return c.baseURL != ""
}

// GetHeartbeat fetches /api/heartbeat.
func (c *Client) GetHeartbeat(ctx context.Context) (*HeartbeatResponse, error) {
	if !c.HasURL() {
		return nil, fmt.Errorf("romm url is not configured")
	}

	url := c.baseURL + "/api/heartbeat"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create heartbeat request: %w", err)
	}

	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("heartbeat HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("heartbeat unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var heartbeat HeartbeatResponse
	if err := json.NewDecoder(resp.Body).Decode(&heartbeat); err != nil {
		return nil, fmt.Errorf("failed to decode heartbeat response: %w", err)
	}

	return &heartbeat, nil
}

// GetPlatforms fetches /api/platforms.
func (c *Client) GetPlatforms(ctx context.Context) ([]PlatformResponse, error) {
	if !c.HasURL() {
		return nil, fmt.Errorf("romm url is not configured")
	}

	url := c.baseURL + "/api/platforms"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create platforms request: %w", err)
	}

	c.setAuth(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("platforms HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("platforms unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var platforms []PlatformResponse
	if err := json.NewDecoder(resp.Body).Decode(&platforms); err != nil {
		return nil, fmt.Errorf("failed to decode platforms response: %w", err)
	}

	return platforms, nil
}

func (c *Client) setAuth(req *http.Request) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	} else if c.user != "" || c.pass != "" {
		req.SetBasicAuth(c.user, c.pass)
	}
}

// Helper functions for type coercion

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func ToFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case bool:
		if val {
			return 1
		}
		return 0
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

func ToBoolFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case bool:
		if val {
			return 1
		}
		return 0
	case float64:
		if val != 0 {
			return 1
		}
		return 0
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		if s == "true" || s == "1" || s == "yes" || s == "enabled" {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// SetTransport sets custom HTTP RoundTripper (primarily for testing).
func (c *Client) SetTransport(rt http.RoundTripper) {
	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}
	c.httpClient.Transport = rt
}
