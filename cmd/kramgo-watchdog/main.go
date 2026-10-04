package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/crystalpoplar/kramgo/internal/app"
)

func main() {
	cfg := app.DefaultWatchdogConfig()

	flag.StringVar(&cfg.OllamaBaseURL, "ollama-url", cfg.OllamaBaseURL, "Ollama base URL")
	flag.StringVar(&cfg.Model, "model", cfg.Model, "Ollama model to validate")
	flag.StringVar(&cfg.ServiceName, "service-name", cfg.ServiceName, "systemd service to restart when the model drifts")
	flag.DurationVar(&cfg.Interval, "interval", cfg.Interval, "seconds between health checks")
	flag.IntVar(&cfg.MaxFailures, "max-failures", cfg.MaxFailures, "number of failed checks before a restart")
	flag.StringVar(&cfg.HealthPrompt, "prompt", cfg.HealthPrompt, "health check prompt to send to the model")
	flag.BoolVar(&cfg.RebuildModel, "rebuild-model", cfg.RebuildModel, "rebuild the model from the Modelfile when a restart is triggered")
	flag.StringVar(&cfg.ModelFilePath, "modelfile", cfg.ModelFilePath, "path to the Modelfile used to recreate the model")
	flag.Parse()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid watchdog config: %v", err)
	}

	log.Printf("starting ollama watchdog for %s on %s", cfg.Model, cfg.OllamaBaseURL)
	if err := cfg.Run(context.Background()); err != nil {
		log.Fatalf("watchdog exited unexpectedly: %v", err)
	}
	_ = time.Second
}
