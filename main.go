package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "time/tzdata"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/romm-exporter/romm-exporter/internal/collector"
	"github.com/romm-exporter/romm-exporter/internal/config"
	"github.com/romm-exporter/romm-exporter/internal/privilege"
	"github.com/romm-exporter/romm-exporter/internal/rommclient"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("[INFO] Starting RomM Prometheus Exporter version %s (commit %s, built %s)", version, commit, date)

	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("[FATAL] Failed to load configuration: %v", err)
	}

	if cfg.TZ != "" {
		loc, err := time.LoadLocation(cfg.TZ)
		if err != nil {
			log.Printf("[WARN] Failed to load requested timezone %q: %v. Using UTC.", cfg.TZ, err)
		} else {
			time.Local = loc
			log.Printf("[INFO] Configured timezone to %s", cfg.TZ)
		}
	}

	// Drop root privileges before binding socket
	if err := privilege.DropPrivileges(cfg.PUID, cfg.PGID); err != nil {
		log.Fatalf("[FATAL] Privilege drop failed: %v", err)
	}

	client := rommclient.NewClient(
		cfg.RommURL,
		cfg.RommUser,
		cfg.RommPass,
		cfg.RommToken,
		cfg.ScrapeTimeout,
	)

	coll := collector.NewCollector(client, cfg.ScrapeTimeout)

	reg := prometheus.NewRegistry()
	reg.MustRegister(coll)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>RomM Prometheus Exporter</title></head>
<body>
<h1>RomM Prometheus Exporter</h1>
<p>Version: %s</p>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`, version)
	})

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[INFO] Exporter listening on %s", cfg.ListenAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] HTTP server listener error: %v", err)
		}
	}()

	<-stop
	log.Printf("[INFO] Shutting down HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] HTTP server shutdown error: %v", err)
	} else {
		log.Printf("[INFO] HTTP server shutdown cleanly.")
	}
}
