package collector

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/romm-exporter/romm-exporter/internal/rommclient"
)

type mockRoundTripper struct {
	fn func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

func TestCollectorSuccess(t *testing.T) {
	mockTransport := &mockRoundTripper{
		fn: func(r *http.Request) (*http.Response, error) {
			var body string
			switch r.URL.Path {
			case "/api/heartbeat":
				body = `{
					"SYSTEM": { "VERSION": "3.8.0" },
					"METADATA_SOURCES": {
						"IGDB_API_ENABLED": true,
						"MOBY_API_ENABLED": false,
						"ANY_SOURCE_ENABLED": true
					}
				}`
			case "/api/platforms":
				body = `[
					{
						"id": 1,
						"slug": "gba",
						"name": "Game Boy Advance",
						"display_name": "GBA",
						"category": "Console",
						"generation": 6,
						"family_name": "Game Boy",
						"family_slug": "gameboy",
						"igdb_id": 24,
						"moby_id": 12,
						"ss_id": null,
						"ra_id": 5,
						"hasheous_id": null,
						"tgdb_id": 4,
						"launchbox_id": 10,
						"is_identified": true,
						"missing_from_fs": false,
						"rom_count": 42,
						"fs_size_bytes": 104857600,
						"firmware_count": 1
					}
				]`
			default:
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(bytes.NewBufferString(""))}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(body)),
				Header:     make(http.Header),
			}, nil
		},
	}

	client := rommclient.NewClient("http://localhost:8080", "user", "pass", "", 5*time.Second)
	client.SetTransport(mockTransport)

	coll := NewCollector(client, 5*time.Second)
	reg := prometheus.NewRegistry()
	reg.MustRegister(coll)

	expectedMetrics := `
# HELP romm_any_metadata_source_enabled 1 if any metadata source is enabled in RomM config, 0 otherwise.
# TYPE romm_any_metadata_source_enabled gauge
romm_any_metadata_source_enabled 1
# HELP romm_auth_ok 1 if authenticated call to RomM API succeeded, 0 otherwise.
# TYPE romm_auth_ok gauge
romm_auth_ok 1
# HELP romm_metadata_source_enabled 1 if metadata source integration is enabled in RomM config, 0 otherwise.
# TYPE romm_metadata_source_enabled gauge
romm_metadata_source_enabled{source="igdb"} 1
romm_metadata_source_enabled{source="moby"} 0
# HELP romm_platform_firmware_count Total firmware count for platform.
# TYPE romm_platform_firmware_count gauge
romm_platform_firmware_count{platform_id="1"} 1
# HELP romm_platform_fs_size_bytes Total filesystem size in bytes for platform.
# TYPE romm_platform_fs_size_bytes gauge
romm_platform_fs_size_bytes{platform_id="1"} 1.048576e+08
# HELP romm_platform_info Info metric for platform metadata in RomM.
# TYPE romm_platform_info gauge
romm_platform_info{category="Console",display_name="GBA",family_name="Game Boy",family_slug="gameboy",generation="6",hasheous_id="",igdb_id="24",is_identified="true",launchbox_id="10",missing_from_fs="false",moby_id="12",name="Game Boy Advance",platform_id="1",ra_id="5",slug="gba",ss_id="",tgdb_id="4"} 1
# HELP romm_platform_rom_count Total ROM count for platform.
# TYPE romm_platform_rom_count gauge
romm_platform_rom_count{platform_id="1"} 42
# HELP romm_up 1 if RomM API heartbeat responds with HTTP 200, 0 otherwise.
# TYPE romm_up gauge
romm_up 1
# HELP romm_version_info RomM version string.
# TYPE romm_version_info gauge
romm_version_info{version="3.8.0"} 1
`

	err := testutil.GatherAndCompare(reg, strings.NewReader(expectedMetrics))
	if err != nil {
		t.Fatalf("unexpected metric output mismatch: %v", err)
	}
}

func TestCollectorPartialFailure(t *testing.T) {
	mockTransport := &mockRoundTripper{
		fn: func(r *http.Request) (*http.Response, error) {
			if r.URL.Path == "/api/heartbeat" {
				body := `{"SYSTEM": {"VERSION": "3.8.0"}}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
				}, nil
			}
			// Platforms fail (500)
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewBufferString("server error")),
			}, nil
		},
	}

	client := rommclient.NewClient("http://localhost:8080", "user", "pass", "", 5*time.Second)
	client.SetTransport(mockTransport)

	coll := NewCollector(client, 5*time.Second)
	reg := prometheus.NewRegistry()
	reg.MustRegister(coll)

	expectedMetrics := `
# HELP romm_auth_ok 1 if authenticated call to RomM API succeeded, 0 otherwise.
# TYPE romm_auth_ok gauge
romm_auth_ok 0
# HELP romm_up 1 if RomM API heartbeat responds with HTTP 200, 0 otherwise.
# TYPE romm_up gauge
romm_up 1
# HELP romm_version_info RomM version string.
# TYPE romm_version_info gauge
romm_version_info{version="3.8.0"} 1
`

	err := testutil.GatherAndCompare(reg, strings.NewReader(expectedMetrics))
	if err != nil {
		t.Fatalf("unexpected metric output mismatch: %v", err)
	}
}
