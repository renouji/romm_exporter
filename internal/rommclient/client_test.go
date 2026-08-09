package rommclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func TestClientHeartbeat(t *testing.T) {
	mockTransport := &mockRoundTripper{
		fn: func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/heartbeat" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer rmm_testtoken" {
				t.Errorf("unexpected auth header: %s", auth)
			}
			body := `{
				"SYSTEM": { "VERSION": "3.8.0" },
				"METADATA_SOURCES": {
					"IGDB_API_ENABLED": true,
					"MOBY_API_ENABLED": false,
					"ANY_SOURCE_ENABLED": true
				}
			}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := NewClient("http://localhost:8080", "", "", "rmm_testtoken", 5*time.Second)
	client.httpClient.Transport = mockTransport

	resp, err := client.GetHeartbeat(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.System.Version != "3.8.0" {
		t.Errorf("expected version 3.8.0, got %s", resp.System.Version)
	}
	if ToBoolFloat64(resp.MetadataSources["IGDB_API_ENABLED"]) != 1 {
		t.Errorf("expected IGDB_API_ENABLED to be 1")
	}
	if ToBoolFloat64(resp.MetadataSources["MOBY_API_ENABLED"]) != 0 {
		t.Errorf("expected MOBY_API_ENABLED to be 0")
	}
}

func TestClientPlatforms(t *testing.T) {
	mockTransport := &mockRoundTripper{
		fn: func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/api/platforms" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user != "admin" || pass != "secret" {
				t.Errorf("invalid basic auth: ok=%v, user=%s", ok, user)
			}
			body := `[
				{
					"id": 1,
					"slug": "gba",
					"name": "Game Boy Advance",
					"display_name": "Game Boy Advance",
					"category": "Console",
					"generation": 6,
					"family_name": "Game Boy",
					"family_slug": "gameboy",
					"rom_count": 42,
					"fs_size_bytes": 104857600,
					"firmware_count": 1
				}
			]`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := NewClient("http://localhost:8080", "admin", "secret", "", 5*time.Second)
	client.httpClient.Transport = mockTransport

	platforms, err := client.GetPlatforms(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(platforms) != 1 {
		t.Fatalf("expected 1 platform, got %d", len(platforms))
	}

	p := platforms[0]
	if ToString(p.ID) != "1" {
		t.Errorf("expected ID 1, got %s", ToString(p.ID))
	}
	if p.Slug != "gba" {
		t.Errorf("expected slug gba, got %s", p.Slug)
	}
	if ToFloat64(p.RomCount) != 42 {
		t.Errorf("expected rom_count 42, got %v", ToFloat64(p.RomCount))
	}
}

func TestCoercionHelpers(t *testing.T) {
	if ToString(nil) != "" {
		t.Errorf("expected empty string for nil")
	}
	if ToString(123) != "123" {
		t.Errorf("expected 123 string")
	}
	if ToString(123.0) != "123" {
		t.Errorf("expected 123 string for 123.0")
	}
	if ToString(true) != "true" {
		t.Errorf("expected true string")
	}

	if ToFloat64("42.5") != 42.5 {
		t.Errorf("expected float 42.5")
	}
	if ToFloat64(nil) != 0 {
		t.Errorf("expected 0 for nil float")
	}

	if ToBoolFloat64(true) != 1 {
		t.Errorf("expected 1 for true")
	}
	if ToBoolFloat64("enabled") != 1 {
		t.Errorf("expected 1 for enabled")
	}
	if ToBoolFloat64(false) != 0 {
		t.Errorf("expected 0 for false")
	}
}
