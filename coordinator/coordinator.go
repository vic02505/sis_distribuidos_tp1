package main

import (
	"log"
	"os"
	"strconv"
	"tp1/coordinator/internal/communications"
)

func main() {

	reducersAmount, err := strconv.Atoi(os.Args[1])

	if err != nil {
		log.Fatal(err)
	}

	fileSplits := os.Args[2:]

	coordinator := communications.NewCoordinator(fileSplits, uint8(reducersAmount))
	coordinator.StartCoordinator()
}
