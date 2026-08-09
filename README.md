# RomM Prometheus Exporter

A lightweight, high-performance Prometheus exporter for [RomM](https://github.com/rommapp/romm) (ROM Manager) written in Go.

## Features

- **Health & Version Monitoring**: Exposes system status (`romm_up`) and RomM release version (`romm_version_info`).
- **Metadata Integration Status**: Tracks which metadata source backends (IGDB, MobyGames, Steam, etc.) are enabled.
- **Platform Analytics**: Comprehensive per-platform metrics including ROM counts, total disk usage (`fs_size_bytes`), firmware count, and rich metadata labels (category, generation, family, external database IDs).
- **Graceful Error Isolation**: Endpoint scraping failures are handled independently—a single failing API endpoint will not crash the exporter or drop metrics for unaffected endpoints.
- **Security First**: Runs unprivileged by default in Docker containers using customizable in-binary `PUID`/`PGID` privilege dropping.
- **Multi-Arch Docker Images**: Native support for `linux/amd64` and `linux/arm64`.

---

## Configuration

The exporter is configured entirely via environment variables:

| Environment Variable | Default | Description |
| :--- | :--- | :--- |
| `LISTEN_ADDR` | `:8585` | Address and port for the HTTP server to listen on. |
| `ROMM_URL` | *(Required)* | Base URL of your RomM instance (e.g. `http://romm.local:8080`). |
| `ROMM_TOKEN` | *(Optional)* | RomM API Bearer Token (format `rmm_...`). Takes precedence over user/pass. |
| `ROMM_USER` | *(Optional)* | Username for RomM Basic Auth. |
| `ROMM_PASS` | *(Optional)* | Password for RomM Basic Auth. |
| `PUID` | `1000` | User ID to run process as (drops root privileges at startup). |
| `PGID` | `1000` | Group ID to run process as. |
| `TZ` | `UTC` | Timezone location for log timestamps. |
| `SCRAPE_TIMEOUT` | `10s` | HTTP timeout duration per RomM API endpoint scrape. |

---

## Quick Start

### Docker Compose

```yaml
version: '3.8'

services:
  romm-exporter:
    image: ghcr.io/romm-exporter/romm-exporter:latest
    container_name: romm-exporter
    restart: unless-stopped
    ports:
      - "8585:8585"
    environment:
      - LISTEN_ADDR=:8585
      - ROMM_URL=http://romm:8080
      - ROMM_TOKEN=rmm_your_bearer_token_here
      - PUID=1000
      - PGID=1000
      - TZ=America/New_York
```

### Docker CLI

```bash
docker run -d \
  --name romm-exporter \
  -p 8585:8585 \
  -e ROMM_URL=http://192.168.1.100:8080 \
  -e ROMM_TOKEN=rmm_your_bearer_token_here \
  ghcr.io/romm-exporter/romm-exporter:latest
```

---

## Prometheus Configuration

Add the following job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'romm'
    scrape_interval: 1m
    scrape_timeout: 10s
    static_configs:
      - targets: ['romm-exporter:8585']
```

---

## Exported Metrics

| Metric Name | Type | Description | Labels |
| :--- | :--- | :--- | :--- |
| `romm_up` | Gauge | `1` if RomM API heartbeat responds with HTTP 200, `0` otherwise. | *(None)* |
| `romm_version_info` | Gauge | `1` info metric showing RomM system version string. | `version` |
| `romm_metadata_source_enabled` | Gauge | `1` if metadata source is enabled, `0` otherwise. | `source` |
| `romm_any_metadata_source_enabled` | Gauge | `1` if any metadata provider is enabled, `0` otherwise. | *(None)* |
| `romm_auth_ok` | Gauge | `1` if authenticated call to `/api/platforms` succeeded, `0` otherwise. | *(None)* |
| `romm_platform_info` | Gauge | `1` info metric with full metadata per platform. | `platform_id`, `slug`, `name`, `display_name`, `category`, `generation`, `family_name`, `family_slug`, `igdb_id`, `moby_id`, `ss_id`, `ra_id`, `hasheous_id`, `tgdb_id`, `launchbox_id`, `is_identified`, `missing_from_fs` |
| `romm_platform_rom_count` | Gauge | Total count of ROMs for the platform. | `platform_id` |
| `romm_platform_fs_size_bytes` | Gauge | Total filesystem disk space in bytes used by platform ROMs. | `platform_id` |
| `romm_platform_firmware_count` | Gauge | Total count of firmware files associated with the platform. | `platform_id` |

---

## Development & Testing

### Running Tests

```bash
go test -v ./...
```

### Building Locally

```bash
go build -ldflags="-s -w -X main.version=v0.1.0-alpha.1" -o romm-exporter .
```

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
