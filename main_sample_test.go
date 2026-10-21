package main

import (
	"GoPower/db"
	"GoPower/models"
	"GoPower/repository"
	"GoPower/server"
	"GoPower/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestSample_RecordGenerationAndDispatch(t *testing.T) {
	database, err := db.InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	stRepo := repository.NewStationRepository(database)
	csRepo := repository.NewConsumerRepository(database)
	gridService := service.NewGridService(database, stRepo, csRepo)

	solar := &models.PowerStation{
		ID:         "st-sol-01",
		Name:       "Shiraz Solar Farm",
		Type:       models.StationSolar,
		CapacityMW: 100.0,
		IsActive:   true,
	}
	_ = stRepo.Create(solar)

	consumer := &models.ConsumerMeter{
		ID:           "cm-ind-01",
		Name:         "Aluminum Smelter",
		Type:         models.ConsumerIndustrial,
		ContractedMW: 60.0,
		IsConnected:  true,
	}
	_ = csRepo.Create(consumer)

	// Record generation output
	if err := gridService.RecordGeneration("st-sol-01", 80.0); err != nil {
		t.Fatalf("failed to record generation: %v", err)
	}

	// Dispatch energy load
	record, err := gridService.DispatchEnergy("st-sol-01", "cm-ind-01", 30.0)
	if err != nil {
		t.Fatalf("failed to dispatch energy: %v", err)
	}

	if record.AmountMW != 30.0 {
		t.Errorf("dispatched amount mismatch: expected 30.0, got %f", record.AmountMW)
	}

	// Fetch grid summary
	summary, err := gridService.GetGridSummary()
	if err != nil {
		t.Fatalf("failed to get grid summary: %v", err)
	}

	if summary.NetBalanceMW != 20.0 { // (80 - 30) generation - 30 demand = 50 - 30 = 20
		t.Errorf("net balance mismatch: expected 20.0, got %f", summary.NetBalanceMW)
	}
}

func TestSample_CreateStationAndHealthEndpoint(t *testing.T) {
	database, err := db.InitDB("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	stRepo := repository.NewStationRepository(database)
	csRepo := repository.NewConsumerRepository(database)
	gridService := service.NewGridService(database, stRepo, csRepo)
	router := server.SetupRouter(gridService, stRepo, csRepo)

	// Test health check endpoint GET /health
	reqHealth, _ := http.NewRequest(http.MethodGet, "/health", nil)
	wHealth := httptest.NewRecorder()
	router.ServeHTTP(wHealth, reqHealth)

	if wHealth.Code != http.StatusOK {
		t.Errorf("health endpoint status mismatch: expected 200, got %d", wHealth.Code)
	}

	// Test create station POST /api/v1/stations
	body := map[string]any{
		"id":          "st-api-01",
		"name":        "Kavir Solar Farm",
		"type":        "SOLAR",
		"capacity_mw": 150.0,
		"is_active":   true,
	}
	bodyBytes, _ := json.Marshal(body)
	reqStation, _ := http.NewRequest(http.MethodPost, "/api/v1/stations", bytes.NewBuffer(bodyBytes))
	reqStation.Header.Set("Content-Type", "application/json")
	wStation := httptest.NewRecorder()
	router.ServeHTTP(wStation, reqStation)

	if wStation.Code != http.StatusCreated {
		t.Errorf("create station status mismatch: expected 201, got %d", wStation.Code)
	}
}
