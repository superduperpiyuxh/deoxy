package main

import "github.com/superduperpiyuxh/deoxy/cmd/deoxy/cmd"

// version is the current release version of deoxy.
// Overridden at build time via -ldflags -X main.version=<ver>.
var version = "0.1.0"

func main() {
	cmd.Execute()
}
