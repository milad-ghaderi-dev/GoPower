package models

import (
	"errors"
	"time"
)

type ConsumerMeter struct {
	ID           string       `gorm:"primaryKey;size:64" json:"id"`
	Name         string       `gorm:"not null;size:128" json:"name"`
	Type         ConsumerType `gorm:"not null;size:32" json:"type"`
	ContractedMW float64      `gorm:"not null" json:"contracted_mw"`
	CurrentLoad  float64      `gorm:"not null;default:0" json:"current_load_mw"`
	IsConnected  bool         `gorm:"not null" json:"is_connected"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

func (c *ConsumerMeter) Validate() error {
	if c.ID == "" {
		return ValidationError{
			Field:  "ID",
			Reason: ErrInvalidConsumerID.Error(),
		}
	}

	if c.Name == "" {
		return ValidationError{
			Field:  "Name",
			Reason: ErrEmptyConsumerName.Error(),
		}
	}

	if c.Type != ConsumerResidential && c.Type != ConsumerCommercial && c.Type != ConsumerIndustrial {
		return ValidationError{
			Field:  "Type",
			Reason: ErrInvalidConsumerType.Error(),
		}
	}

	if c.ContractedMW <= 0 {
		return ValidationError{
			Field:  "ContractedMW",
			Reason: ErrInvalidContractedMW.Error(),
		}
	}

	if c.CurrentLoad < 0 || c.ContractedMW < c.CurrentLoad {
		return ValidationError{
			Field:  "CurrentLoad",
			Reason: errors.New("current load must be a float between 0 and contracted mw").Error(),
		}
	}

	return nil
}
