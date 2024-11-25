// Package main is the main binary for Devsecdb service.
package main

import (
	"os"

	"github.com/khulnasoft/devsecdb/backend/bin/server/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
