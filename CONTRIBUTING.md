Contributing to RomM Prometheus Exporter
Thank you for your interest in contributing to RomM Prometheus Exporter! Whether you are reporting bugs, proposing new metrics, improving documentation, or building Grafana dashboards, all contributions are welcome.

🚀 Quick Start
Fork the Repository: Create a personal fork of romm_exporter on GitHub.
Clone & Set Up:
git clone https://github.com/<your-username>/romm_exporter.git
cd romm_exporter
Prerequisites:
Go 1.22+
Docker (optional, for container builds)
🛠️ Local Development & Testing
Run all unit tests before submitting changes:

# Run unit tests across all packages
go test -v ./...

# Build local binary
go build -o romm-exporter .

# Test locally against a RomM instance
ROMM_URL="http://<romm-host>:8080" ./romm-exporter
📂 Project Structure
main.go: Application entrypoint, HTTP server setup, and graceful shutdown handling.
internal/config: Environment variable loading and validation.
internal/privilege: Cross-platform privilege dropping (PUID/PGID).
internal/rommclient: HTTP client handling RomM API endpoints (/api/heartbeat, /api/platforms).
internal/collector: Prometheus collector exposing RomM metrics.
✏️ Development Workflow
Create a feature branch:
git checkout -b feat/my-new-metric
# or
git checkout -b fix/issue-description
Make your changes and add tests where appropriate.
Ensure go test ./... passes cleanly.
Commit your changes using clear commit messages (e.g. feat: add per-user stats collector or fix: handle null values in platform response).
Push your branch and open a Pull Request against main.
📊 Grafana Dashboards
Have an awesome Grafana dashboard template?

Feel free to share exported panel/dashboard JSON in issues or PRs so we can feature them in the community documentation or repository!
🐛 Reporting Bugs & Issues
When reporting an issue, please include:

Exporter version (./romm-exporter --version or log output)
RomM version
Relevant log output from romm-exporter
Expected vs. actual behavior