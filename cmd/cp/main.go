package main

import (
	"log"

	"github.com/google/uuid"
)

type CPConfig struct {
	cpID uuid.UUID
}

func loadCPConfig() *CPConfig {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	cpConfig := CPConfig{
		cpID: id,
	}

	return &cpConfig
}

func main() {
	cpConfig := loadCPConfig()

	log.Printf("info: Starting CapabilityProvider(cpID: %v)", cpConfig.cpID)

	StartAPIServer()
}
