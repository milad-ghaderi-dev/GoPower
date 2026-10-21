package repository

import (
	"GoPower/models"

	"gorm.io/gorm"
)

type StationRepository interface {
	Create(station *models.PowerStation) error
	GetByID(id string) (*models.PowerStation, error)
	ListAll() ([]models.PowerStation, error)
	Update(station *models.PowerStation) error
	Delete(id string) error
	GetByType(stationType models.StationType) ([]models.PowerStation, error)
}

type GormStationRepository struct {
	db *gorm.DB
}

func NewStationRepository(db *gorm.DB) *GormStationRepository {
	return &GormStationRepository{db: db}
}

func (r *GormStationRepository) Create(station *models.PowerStation) error {
	if err := station.Validate(); err != nil {
		return err
	}
	result := r.db.Create(station)
	if result.Error != nil {
		return models.ErrDuplicateStationID
	}
	return nil
}

func (r *GormStationRepository) GetByID(id string) (*models.PowerStation, error) {
	if id == "" {
		return nil, models.ErrInvalidStationID
	}
	var station models.PowerStation
	if result := r.db.First(&station, "id = ?", id); result.Error != nil {
		return nil, models.ErrStationNotFound
	}
	return &station, nil
}

func (r *GormStationRepository) ListAll() ([]models.PowerStation, error) {
	stations := make([]models.PowerStation, 0)
	if err := r.db.Order("created_at asc").Find(&stations).Error; err != nil {
		return nil, err
	}
	return stations, nil
}

func (r *GormStationRepository) Update(station *models.PowerStation) error {
	if err := station.Validate(); err != nil {
		return err
	}
	result := r.db.Model(&models.PowerStation{}).Where("id = ?", station.ID).Select("*").Updates(station)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrStationNotFound
	}
	return nil
}

func (r *GormStationRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&models.PowerStation{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrStationNotFound
	}
	return nil
}

func (r *GormStationRepository) GetByType(stationType models.StationType) ([]models.PowerStation, error) {
	stations := make([]models.PowerStation, 0)
	if err := r.db.Model(&models.PowerStation{}).Where("type = ?", stationType).Find(&stations).Error; err != nil {
		return nil, err
	}
	return stations, nil
}
