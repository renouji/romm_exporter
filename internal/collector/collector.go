package collector

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/romm-exporter/romm-exporter/internal/rommclient"
)

type Collector struct {
	client        *rommclient.Client
	scrapeTimeout time.Duration

	rommUp                        *prometheus.Desc
	rommVersionInfo               *prometheus.Desc
	rommMetadataSourceEnabled     *prometheus.Desc
	rommAnyMetadataSourceEnabled  *prometheus.Desc
	rommAuthOK                    *prometheus.Desc
	rommPlatformInfo              *prometheus.Desc
	rommPlatformRomCount          *prometheus.Desc
	rommPlatformFSSizeBytes       *prometheus.Desc
	rommPlatformFirmwareCount     *prometheus.Desc
}

func NewCollector(client *rommclient.Client, scrapeTimeout time.Duration) *Collector {
	return &Collector{
		client:        client,
		scrapeTimeout: scrapeTimeout,

		rommUp: prometheus.NewDesc(
			"romm_up",
			"1 if RomM API heartbeat responds with HTTP 200, 0 otherwise.",
			nil, nil,
		),
		rommVersionInfo: prometheus.NewDesc(
			"romm_version_info",
			"RomM version string.",
			[]string{"version"}, nil,
		),
		rommMetadataSourceEnabled: prometheus.NewDesc(
			"romm_metadata_source_enabled",
			"1 if metadata source integration is enabled in RomM config, 0 otherwise.",
			[]string{"source"}, nil,
		),
		rommAnyMetadataSourceEnabled: prometheus.NewDesc(
			"romm_any_metadata_source_enabled",
			"1 if any metadata source is enabled in RomM config, 0 otherwise.",
			nil, nil,
		),
		rommAuthOK: prometheus.NewDesc(
			"romm_auth_ok",
			"1 if authenticated call to RomM API succeeded, 0 otherwise.",
			nil, nil,
		),
		rommPlatformInfo: prometheus.NewDesc(
			"romm_platform_info",
			"Info metric for platform metadata in RomM.",
			[]string{
				"platform_id",
				"slug",
				"name",
				"display_name",
				"category",
				"generation",
				"family_name",
				"family_slug",
				"igdb_id",
				"moby_id",
				"ss_id",
				"ra_id",
				"hasheous_id",
				"tgdb_id",
				"launchbox_id",
				"is_identified",
				"missing_from_fs",
			}, nil,
		),
		rommPlatformRomCount: prometheus.NewDesc(
			"romm_platform_rom_count",
			"Total ROM count for platform.",
			[]string{"platform_id"}, nil,
		),
		rommPlatformFSSizeBytes: prometheus.NewDesc(
			"romm_platform_fs_size_bytes",
			"Total filesystem size in bytes for platform.",
			[]string{"platform_id"}, nil,
		),
		rommPlatformFirmwareCount: prometheus.NewDesc(
			"romm_platform_firmware_count",
			"Total firmware count for platform.",
			[]string{"platform_id"}, nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.rommUp
	ch <- c.rommVersionInfo
	ch <- c.rommMetadataSourceEnabled
	ch <- c.rommAnyMetadataSourceEnabled
	ch <- c.rommAuthOK
	ch <- c.rommPlatformInfo
	ch <- c.rommPlatformRomCount
	ch <- c.rommPlatformFSSizeBytes
	ch <- c.rommPlatformFirmwareCount
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	if !c.client.HasURL() {
		ch <- prometheus.MustNewConstMetric(c.rommUp, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.rommAuthOK, prometheus.GaugeValue, 0)
		return
	}

	// 1. Scrape Heartbeat
	c.collectHeartbeat(ch)

	// 2. Scrape Platforms
	c.collectPlatforms(ch)
}

func (c *Collector) collectHeartbeat(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.scrapeTimeout)
	defer cancel()

	heartbeat, err := c.client.GetHeartbeat(ctx)
	if err != nil {
		log.Printf("[WARN] Failed to scrape heartbeat endpoint: %v", err)
		ch <- prometheus.MustNewConstMetric(c.rommUp, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.rommUp, prometheus.GaugeValue, 1)

	if heartbeat.System.Version != "" {
		ch <- prometheus.MustNewConstMetric(c.rommVersionInfo, prometheus.GaugeValue, 1, heartbeat.System.Version)
	}

	for k, v := range heartbeat.MetadataSources {
		upperK := strings.ToUpper(k)
		if upperK == "ANY_SOURCE_ENABLED" {
			ch <- prometheus.MustNewConstMetric(c.rommAnyMetadataSourceEnabled, prometheus.GaugeValue, rommclient.ToBoolFloat64(v))
		} else if strings.HasSuffix(upperK, "_API_ENABLED") {
			source := strings.ToLower(strings.TrimSuffix(upperK, "_API_ENABLED"))
			ch <- prometheus.MustNewConstMetric(c.rommMetadataSourceEnabled, prometheus.GaugeValue, rommclient.ToBoolFloat64(v), source)
		}
	}
}

func (c *Collector) collectPlatforms(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.scrapeTimeout)
	defer cancel()

	platforms, err := c.client.GetPlatforms(ctx)
	if err != nil {
		log.Printf("[WARN] Failed to scrape platforms endpoint: %v", err)
		ch <- prometheus.MustNewConstMetric(c.rommAuthOK, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.rommAuthOK, prometheus.GaugeValue, 1)

	for _, p := range platforms {
		platformID := rommclient.ToString(p.ID)
		if platformID == "" {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.rommPlatformInfo,
			prometheus.GaugeValue,
			1,
			platformID,
			p.Slug,
			p.Name,
			p.DisplayName,
			p.Category,
			rommclient.ToString(p.Generation),
			p.FamilyName,
			p.FamilySlug,
			rommclient.ToString(p.IgdbID),
			rommclient.ToString(p.MobyID),
			rommclient.ToString(p.SsID),
			rommclient.ToString(p.RaID),
			rommclient.ToString(p.HasheousID),
			rommclient.ToString(p.TgdbID),
			rommclient.ToString(p.LaunchboxID),
			rommclient.ToString(p.IsIdentified),
			rommclient.ToString(p.MissingFromFS),
		)

		ch <- prometheus.MustNewConstMetric(
			c.rommPlatformRomCount,
			prometheus.GaugeValue,
			rommclient.ToFloat64(p.RomCount),
			platformID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.rommPlatformFSSizeBytes,
			prometheus.GaugeValue,
			rommclient.ToFloat64(p.FSSizeBytes),
			platformID,
		)

		ch <- prometheus.MustNewConstMetric(
			c.rommPlatformFirmwareCount,
			prometheus.GaugeValue,
			rommclient.ToFloat64(p.FirmwareCount),
			platformID,
		)
	}
}
