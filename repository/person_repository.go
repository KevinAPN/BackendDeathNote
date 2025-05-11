package repository

import (
	"backend-avanzada/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type PeopleRepository struct {
	db *gorm.DB
}

func NewPeopleRepository(db *gorm.DB) *PeopleRepository {
	return &PeopleRepository{
		db: db,
	}
}

func (p *PeopleRepository) FindAll() ([]*models.Person, error) {
	var people []*models.Person
	err := p.db.Find(&people).Error
	if err != nil {
		return nil, err
	}
	return people, nil
}

func (p *PeopleRepository) Save(data *models.Person) (*models.Person, error) {
	err := p.db.Save(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *PeopleRepository) FindById(id int) (*models.Person, error) {
	var person models.Person
	err := p.db.Where("id = ?", id).First(&person).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &person, nil
}

func (p *PeopleRepository) Delete(data *models.Person) error {
	err := p.db.Delete(data).Error
	if err != nil {
		return err
	}
	return nil
}

// MarkHeartAttack asigna la causa y marca la hora de muerte
func (p *PeopleRepository) MarkHeartAttack(id uint) error {
	now := time.Now()
	updates := map[string]interface{}{
		"cause":      "ataque al corazón",
		"death_time": now,
	}
	return p.db.Model(&models.Person{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// Añade la causa sin marcar la muerte
func (p *PeopleRepository) AddCause(id uint, cause string) error {
	return p.db.Model(&models.Person{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"cause": cause,
		}).Error
}

// Añade los detalles sin marcar la muerte
func (p *PeopleRepository) AddDetails(id uint, details string) error {
	return p.db.Model(&models.Person{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"details": details,
		}).Error
}

// Marca la muerte definitiva (está en cola tras causa o detalles)
func (p *PeopleRepository) MarkDeath(id uint) error {
	now := time.Now()
	return p.db.Model(&models.Person{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"death_time": now,
		}).Error
}
