package repository

import (
	"errors"

	"github.com/erick_nilson/mysql_auto_backup/internal/models"
	"gorm.io/gorm"
)

type ConnectionRepository struct {
	db *gorm.DB
}

func NewConnectionRepository(db *gorm.DB) *ConnectionRepository {
	return &ConnectionRepository{db: db}
}

func (r *ConnectionRepository) Create(c *models.Connection) error {
	return r.db.Create(c).Error
}

func (r *ConnectionRepository) Update(c *models.Connection) error {
	return r.db.Save(c).Error
}

func (r *ConnectionRepository) FindByID(id uint64) (*models.Connection, error) {
	var c models.Connection
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ConnectionRepository) FindByIDs(ids []uint64) ([]models.Connection, error) {
	var out []models.Connection
	if err := r.db.Where("id IN ?", ids).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ConnectionRepository) List() ([]models.Connection, error) {
	var out []models.Connection
	if err := r.db.Order("id DESC").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ConnectionRepository) ListActive() ([]models.Connection, error) {
	var out []models.Connection
	if err := r.db.Where("is_active = ?", true).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ConnectionRepository) Delete(id uint64) error {
	res := r.db.Delete(&models.Connection{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("conexão não encontrada")
	}
	return nil
}
