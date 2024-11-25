//go:build !minidemo

package server

import (
	// Drivers under linux.
	_ "github.com/khulnasoft/devsecdb/backend/plugin/db/obo"
)
