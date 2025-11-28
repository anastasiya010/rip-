package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"metoda/internal/app/ds"
)

// GetAllMedications - получение всех действующих лекарств
func (r *Repository) GetAllMedications() ([]ds.Medication, error) {
	var medications []ds.Medication
	err := r.db.Where("is_deleted = ?", false).Find(&medications).Error
	if err != nil {
		return nil, err
	}
	return medications, nil
}

// SearchMedicationsByName - поиск лекарств по названию
func (r *Repository) SearchMedicationsByName(name string) ([]ds.Medication, error) {
	var medications []ds.Medication
	err := r.db.Where("name ILIKE ? AND is_deleted = ?", "%"+name+"%", false).Find(&medications).Error
	if err != nil {
		return nil, err
	}
	return medications, nil
}

// GetMedicationByID - получение лекарства по ID (используем курсор как в методичке)
func (r *Repository) GetMedicationByID(id uint) (*ds.Medication, error) {
	query := "SELECT id, name, description, is_deleted, image_url, category, manufacturer, short_info FROM medications WHERE id = $1 AND is_deleted = false"
	// Создание курсора (строковый указатель)
	row := r.db.Raw(query, id).Row()
	// Создание объекта для хранения данных
	medication := &ds.Medication{}
	// Сканирование строки в структуру
	err := row.Scan(
		&medication.ID,
		&medication.Name,
		&medication.Description,
		&medication.IsDeleted,
		&medication.ImageURL,
		&medication.Category,
		&medication.Manufacturer,
		&medication.ShortInfo,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Возвращаем nil, если записи нет
		}
		return nil, err
	}
	return medication, nil
}

// DeleteMedication - логическое удаление лекарства
func (r *Repository) DeleteMedication(medicationID uint) error {
	err := r.db.Model(&ds.Medication{}).Where("id = ?", medicationID).UpdateColumn("is_deleted", true).Error
	if err != nil {
		return fmt.Errorf("ошибка при удалении лекарства с id %d: %w", medicationID, err)
	}
	return nil
}

