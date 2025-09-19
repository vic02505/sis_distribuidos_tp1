package main

import (
	"os"
	"tp1/coordinator/internal/communications"
)

func main() {
	fileSplits := os.Args[1:]

	coordinator := communications.NewCoordinator(fileSplits, 4)
	coordinator.StartCoordinator()
}
