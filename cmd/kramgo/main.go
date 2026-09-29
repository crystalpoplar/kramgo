package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/crystalpoplar/kramgo/internal/app"
)

func run(w io.Writer, jsonPath string) error {
	if err := app.Run(w); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	data, err := app.GetJSONFile(jsonPath)
	if err != nil {
		return fmt.Errorf("read json file: %w", err)
	}

	if _, err := fmt.Fprintln(w, string(data)); err != nil {
		return fmt.Errorf("write output: %w", err)
	}

	return nil
}

func main() {
	ollamaBaseURL := flag.String("ollama-url", "http://127.0.0.1:11434", "Ollama base URL")
	listenAddr := flag.String("listen", "0.0.0.0:8080", "Address the API should bind to")
	startServer := flag.Bool("serve", false, "Start the local network Ollama proxy service")
	flag.Parse()

	if *startServer {
		service := app.NewService(*ollamaBaseURL, *listenAddr)
		log.Printf("starting Ollama proxy on %s -> %s", service.ListenAddr, service.OllamaBaseURL)
		if err := service.ListenAndServe(); err != nil {
			log.Fatalf("server error: %v", err)
		}
		return
	}

	if err := run(os.Stdout, "kramgo.json"); err != nil {
		log.Fatalf("%v", err)
	}
}
