package main

import (
	"log"

	"github.com/google/uuid"
)

type CPConfig struct {
	cpID uuid.UUID
}

func loadCPConfig() CPConfig {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	cpConfig := CPConfig{
		cpID: id,
	}

	return cpConfig
}

var config CPConfig = loadCPConfig()

func main() {
	log.Printf("info: Starting CapabilityProvider(cpID: %v)", config.cpID)

	StartAPIServer()
}
