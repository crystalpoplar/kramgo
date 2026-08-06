package main

import (
	"os"

	"github.com/crystalpoplar/kramgo/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
