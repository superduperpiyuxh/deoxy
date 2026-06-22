package main

import "fmt"

// version is the current release version of deoxy.
// Overridden at build time via -ldflags -X main.version=<ver>.
var version = "0.1.0"

func main() {
	fmt.Printf("deoxy v%s\n", version)
}
