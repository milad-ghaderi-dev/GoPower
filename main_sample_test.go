package main

import (
	"GoPower/db"
	"GoPower/models"
	"GoPower/repository"
	"testing"
)

func TestSample_CreateAndGetStation(t *testing.T) {
	database, err := db.InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	repo := repository.NewStationRepository(database)

	station := &models.PowerStation{
		ID:         "st-sol-01",
		Name:       "Shiraz Solar Farm",
		Type:       models.StationSolar,
		CapacityMW: 120.5,
		IsActive:   true,
	}

	if err := repo.Create(station); err != nil {
		t.Fatalf("failed to create station: %v", err)
	}

	fetched, err := repo.GetByID("st-sol-01")
	if err != nil {
		t.Fatalf("failed to get station by ID: %v", err)
	}

	if fetched.Name != "Shiraz Solar Farm" {
		t.Errorf("station name mismatch: expected 'Shiraz Solar Farm', got '%s'", fetched.Name)
	}
}

func TestSample_CreateAndGetConsumer(t *testing.T) {
	database, err := db.InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	repo := repository.NewConsumerRepository(database)

	consumer := &models.ConsumerMeter{
		ID:           "cm-ind-01",
		Name:         "Tehran Steel Works",
		Type:         models.ConsumerIndustrial,
		ContractedMW: 45.0,
		IsConnected:  true,
	}

	if err := repo.Create(consumer); err != nil {
		t.Fatalf("failed to create consumer meter: %v", err)
	}

	fetched, err := repo.GetByID("cm-ind-01")
	if err != nil {
		t.Fatalf("failed to get consumer meter by ID: %v", err)
	}

	if fetched.ContractedMW != 45.0 {
		t.Errorf("contracted MW mismatch: expected 45.0, got %f", fetched.ContractedMW)
	}
}
