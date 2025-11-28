package repository

import (
	"database/sql"
	"errors"
	"strings"
	"metoda/internal/app/ds"
	"time"
	"gorm.io/gorm"
)

// GetDraftPrescription - получение черновика заявки пользователя
func (r *Repository) GetDraftPrescription(creatorID uint) (*ds.Prescription, error) {
	var prescription ds.Prescription
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").First(&prescription).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

// CreateDraftPrescription - создание новой заявки в статусе черновик
func (r *Repository) CreateDraftPrescription(creatorID uint) (*ds.Prescription, error) {
	// Сначала еще раз проверяем, нет ли уже черновика (на случай race condition)
	existing, err := r.GetDraftPrescription(creatorID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	prescription := &ds.Prescription{
		Status:     "черновик",
		DateCreate: time.Now(),
		CreatorID:  creatorID,
	}
	err = r.db.Create(prescription).Error
	if err != nil {
		// Если ошибка из-за уникального индекса (уже есть черновик), пытаемся найти его
		if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "duplicate key") {
			existing, findErr := r.GetDraftPrescription(creatorID)
			if findErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}
	return prescription, nil
}

// GetPrescriptionByID - получение заявки по ID (без удаленных)
func (r *Repository) GetPrescriptionByID(id uint) (*ds.Prescription, error) {
	var prescription ds.Prescription
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("id = ? AND status != ?", id, "удалён").First(&prescription).Error
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

// GetPrescriptionMedications - получение лекарств в заявке
func (r *Repository) GetPrescriptionMedications(prescriptionID uint) ([]ds.PrescriptionMedication, error) {
	var prescriptionMedications []ds.PrescriptionMedication
	err := r.db.Preload("Medication").
		Where("prescription_id = ?", prescriptionID).
		Order("order_number ASC").
		Find(&prescriptionMedications).Error
	if err != nil {
		return nil, err
	}
	return prescriptionMedications, nil
}

// AddMedicationToPrescription - добавление лекарства в заявку через ORM
func (r *Repository) AddMedicationToPrescription(prescriptionID, medicationID uint, quantity int) error {
	// Проверяем, не добавлено ли уже это лекарство
	var existing ds.PrescriptionMedication
	err := r.db.Where("prescription_id = ? AND medication_id = ?", prescriptionID, medicationID).
		First(&existing).Error
	if err == nil {
		// Если уже есть, увеличиваем количество
		existing.Quantity += quantity
		return r.db.Save(&existing).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Получаем максимальный порядковый номер
	var maxOrder int
	r.db.Model(&ds.PrescriptionMedication{}).
		Where("prescription_id = ?", prescriptionID).
		Select("COALESCE(MAX(order_number), 0)").
		Scan(&maxOrder)

	prescriptionMedication := &ds.PrescriptionMedication{
		PrescriptionID: prescriptionID,
		MedicationID:   medicationID,
		Quantity:       quantity,
		OrderNumber:    maxOrder + 1,
		IsMain:         false,
	}
	return r.db.Create(prescriptionMedication).Error
}

// GetCartCount - получение количества лекарств в черновике заявки
func (r *Repository) GetCartCount(creatorID uint) int64 {
	var prescriptionID uint
	var count int64

	// пока что мы захардкодили id создателя заявки, в последующем вы сделаете авторизацию и будете получать его из JWT
	err := r.db.Model(&ds.Prescription{}).
		Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("id").First(&prescriptionID).Error
	if err != nil {
		return 0
	}

	err = r.db.Model(&ds.PrescriptionMedication{}).
		Where("prescription_id = ?", prescriptionID).Count(&count).Error
	if err != nil {
		return 0
	}
	return count
}

// GetDraftPrescriptionID - получение ID черновика заявки
func (r *Repository) GetDraftPrescriptionID(creatorID uint) *uint {
	var prescription ds.Prescription
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "черновик").
		Select("id").First(&prescription).Error
	if err != nil {
		return nil
	}
	return &prescription.ID
}

// DeletePrescription - логическое удаление заявки через SQL UPDATE (без ORM)
func (r *Repository) DeletePrescription(prescriptionID uint) error {
	query := "UPDATE prescriptions SET status = $1 WHERE id = $2"
	result := r.db.Exec(query, "удалён", prescriptionID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

