package main

import (
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
	if err := run(os.Stdout, "kramgo.json"); err != nil {
		log.Fatalf("%v", err)
	}
}
