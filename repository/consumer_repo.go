package repository

import (
	"GoPower/models"

	"gorm.io/gorm"
)

type ConsumerRepository interface {
	Create(consumer *models.ConsumerMeter) error
	GetByID(id string) (*models.ConsumerMeter, error)
	ListAll() ([]models.ConsumerMeter, error)
	Update(consumer *models.ConsumerMeter) error
	Delete(id string) error
}

type GormConsumerRepository struct {
	db *gorm.DB
}

func NewConsumerRepository(db *gorm.DB) *GormConsumerRepository {
	return &GormConsumerRepository{db: db}
}

func (r *GormConsumerRepository) Create(consumer *models.ConsumerMeter) error {
	if err := consumer.Validate(); err != nil {
		return err
	}
	if result := r.db.Create(consumer); result.Error != nil {
		return models.ErrDuplicateConsumerID
	}
	return nil
}

func (r *GormConsumerRepository) GetByID(id string) (*models.ConsumerMeter, error) {
	if id == "" {
		return nil, models.ErrInvalidConsumerID
	}
	var consumer models.ConsumerMeter
	result := r.db.First(&consumer, "id = ?", id)
	if result.Error != nil {
		return nil, models.ErrConsumerNotFound
	}
	return &consumer, nil
}

func (r *GormConsumerRepository) ListAll() ([]models.ConsumerMeter, error) {
	consumers := make([]models.ConsumerMeter, 0)
	if err := r.db.Order("created_at asc").Find(&consumers).Error; err != nil {
		return nil, err
	}
	return consumers, nil
}

func (r *GormConsumerRepository) Update(consumer *models.ConsumerMeter) error {
	if err := consumer.Validate(); err != nil {
		return err
	}
	result := r.db.Model(&models.ConsumerMeter{}).Where("id = ?", consumer.ID).Select("*").Updates(consumer)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrConsumerNotFound
	}

	return nil
}

func (r *GormConsumerRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(&models.ConsumerMeter{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return models.ErrConsumerNotFound
	}
	return nil
}
