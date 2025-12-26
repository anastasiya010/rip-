package repository

import (
	"metoda/internal/app/ds"

	"gorm.io/gorm"
)

// DeletePrescriptionMedication - удаление записи м-м из заявки (без PK м-м)
func (r *Repository) DeletePrescriptionMedication(prescriptionID, medicationID uint) error {
	result := r.db.Where("prescription_id = ? AND medication_id = ?", prescriptionID, medicationID).
		Delete(&ds.PrescriptionMedication{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// UpdatePrescriptionMedication - изменение порядка/значения в м-м (без PK м-м)
func (r *Repository) UpdatePrescriptionMedication(prescriptionID, medicationID uint, orderNumber *int, dosageInstruction *string, isMain *bool) error {
	var pm ds.PrescriptionMedication
	err := r.db.Where("prescription_id = ? AND medication_id = ?", prescriptionID, medicationID).
		First(&pm).Error
	if err != nil {
		return err
	}

	// Обновляем только переданные поля
	if orderNumber != nil {
		pm.OrderNumber = *orderNumber
	}
	if dosageInstruction != nil {
		pm.DosageInstruction = *dosageInstruction
	}
	if isMain != nil {
		pm.IsMain = *isMain
	}

	return r.db.Save(&pm).Error
}
