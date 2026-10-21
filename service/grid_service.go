package service

import (
	"GoPower/models"
	"GoPower/repository"
	"sync"
	"time"

	"gorm.io/gorm"
)

type GridService struct {
	db           *gorm.DB
	stationRepo  repository.StationRepository
	consumerRepo repository.ConsumerRepository
	mu           sync.Mutex
}

func NewGridService(db *gorm.DB, stationRepo repository.StationRepository, consumerRepo repository.ConsumerRepository) *GridService {
	return &GridService{
		db:           db,
		stationRepo:  stationRepo,
		consumerRepo: consumerRepo,
	}
}

func (s *GridService) RecordGeneration(stationID string, outputMW float64) error {
	if outputMW < 0 {
		return models.ErrInvalidAmount
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return err
	}
	if !station.IsActive {
		return models.ErrStationInactive
	}
	if station.CapacityMW < outputMW {
		return models.ErrStationOverload
	}
	station.CurrentOutput = outputMW
	if err := s.stationRepo.Update(station); err != nil {
		return err
	}
	return nil
}

func (s *GridService) ChargeBattery(stationID string, amountMWh float64) error {
	if amountMWh <= 0 {
		return models.ErrInvalidAmount
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return err
	}
	if station.Type != models.StationBattery {
		return models.ErrInvalidStationType
	}
	if station.IsActive == false {
		return models.ErrStationInactive
	}
	if (station.StoredEnergy + amountMWh) > station.MaxStorage {
		return models.ErrStationOverload
	}
	station.StoredEnergy += amountMWh
	if err := s.stationRepo.Update(station); err != nil {
		return err
	}
	return nil
}

func (s *GridService) DispatchEnergy(stationID, consumerID string, amountMW float64) (*models.DispatchRecord, error) {
	if amountMW <= 0 {
		return nil, models.ErrInvalidAmount
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	station, err := s.stationRepo.GetByID(stationID)
	if err != nil {
		return nil, err
	}

	if !station.IsActive {
		return nil, models.ErrStationInactive
	}
	consumer, err := s.consumerRepo.GetByID(consumerID)
	if err != nil {
		return nil, err
	}

	if !consumer.IsConnected {
		return nil, models.ErrConsumerDisconnected
	}

	if station.Type == models.StationBattery {
		if station.StoredEnergy < amountMW {
			return nil, models.ErrInsufficientEnergy
		}
		station.StoredEnergy -= amountMW
	} else {
		if station.CurrentOutput < amountMW {
			return nil, models.ErrInsufficientEnergy
		}
		station.CurrentOutput -= amountMW
	}
	if station.Type != models.StationBattery && (station.CurrentOutput+amountMW) > station.CapacityMW {
		return nil, models.ErrStationOverload
	}

	if consumer.CurrentLoad+amountMW > consumer.ContractedMW {
		return nil, models.ErrStationOverload
	}
	consumer.CurrentLoad += amountMW

	var dispatch models.DispatchRecord
	err = s.db.Transaction(func(tx *gorm.DB) error {
		rConsumer := tx.Model(&models.ConsumerMeter{}).Where("id = ?", consumerID).Select("*").Updates(consumer)
		if rConsumer.Error != nil {
			return rConsumer.Error
		}
		if rConsumer.RowsAffected == 0 {
			return models.ErrConsumerNotFound
		}
		rStation := tx.Model(&models.PowerStation{}).Where("id = ?", stationID).Select("*").Updates(station)
		if rStation.Error != nil {
			return rStation.Error
		}
		if rStation.RowsAffected == 0 {
			return models.ErrStationNotFound
		}
		dispatch = models.DispatchRecord{
			StationID:   stationID,
			ConsumerID:  consumerID,
			AmountMW:    amountMW,
			DeliveredAt: time.Now(),
		}
		rDispatch := tx.Model(&models.DispatchRecord{}).Create(&dispatch)
		if rDispatch.Error != nil {
			return rDispatch.Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &dispatch, nil
}

func (s *GridService) GetGridSummary() (*models.GridSummary, error) {
	stations, err := s.stationRepo.ListAll()
	if err != nil {
		return nil, err
	}
	consumers, err := s.consumerRepo.ListAll()
	if err != nil {
		return nil, err
	}
	var summary models.GridSummary
	summary.StationBreakdown = make(map[models.StationType]int)
	for _, s := range stations {
		if s.IsActive {
			summary.TotalStations++
			summary.TotalGenerationMW += s.CurrentOutput
		}
		summary.StationBreakdown[s.Type]++
		summary.TotalCapacityMW += s.CapacityMW

	}
	for _, c := range consumers {
		if c.IsConnected {
			summary.TotalConsumers++
			summary.TotalDemandMW += c.CurrentLoad
		}

	}
	summary.NetBalanceMW = summary.TotalGenerationMW - summary.TotalDemandMW
	if summary.NetBalanceMW >= 0 {
		summary.Status = models.StatusNormal
	} else if summary.NetBalanceMW >= -50 {
		summary.Status = models.StatusWarning
	} else {
		summary.Status = models.StatusCritical
	}
	return &summary, nil
}
