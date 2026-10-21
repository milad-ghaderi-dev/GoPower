package models

import "time"

type PowerStation struct {
	ID            string      `gorm:"primaryKey;size:64" json:"id"`
	Name          string      `gorm:"not null;size:128" json:"name"`
	Type          StationType `gorm:"not null;size:32" json:"type"`
	CapacityMW    float64     `gorm:"not null" json:"capacity_mw"`
	CurrentOutput float64     `gorm:"not null;default:0" json:"current_output_mw"`
	StoredEnergy  float64     `gorm:"not null;default:0" json:"stored_energy_mwh"`
	MaxStorage    float64     `gorm:"not null;default:0" json:"max_storage_mwh"`
	IsActive      bool        `gorm:"not null" json:"is_active"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

func (s *PowerStation) Validate() error {
	if s.ID == "" {
		return ValidationError{
			Field:  "ID",
			Reason: ErrInvalidStationID.Error(),
		}
	}

	if s.Name == "" {
		return ValidationError{
			Field:  "Name",
			Reason: ErrEmptyStationName.Error(),
		}
	}

	if s.Type != StationWind && s.Type != StationBattery && s.Type != StationSolar {
		return ValidationError{
			Field:  "Type",
			Reason: ErrInvalidStationType.Error(),
		}
	}

	if s.Type == StationBattery {
		if s.MaxStorage <= 0 {
			return ValidationError{
				Field:  "MaxStorage",
				Reason: "max storage must be greater than 0",
			}
		}
		if s.StoredEnergy < 0 || s.StoredEnergy > s.MaxStorage {
			return ValidationError{
				Field:  "StoredEnergy",
				Reason: "stored energy must be a float greater than 0 and less than max storage",
			}
		}
	}

	if s.CapacityMW <= 0 {
		return ValidationError{
			Field:  "CapacityMW",
			Reason: ErrInvalidCapacity.Error(),
		}
	}

	if s.CurrentOutput < 0 {
		return ValidationError{
			Field:  "CurrentOutput",
			Reason: "current output cannot be negative",
		}
	}

	if s.CapacityMW < s.CurrentOutput {
		return ValidationError{
			Field:  "CurrentOutput",
			Reason: ErrStationOverload.Error(),
		}
	}

	return nil
}

