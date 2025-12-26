package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"metoda/internal/app/ds"
	"strings"
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

	// Если данные ребенка не установлены, устанавливаем их автоматически
	if prescription.PatientName == "" || prescription.PatientAge == 0 || prescription.PatientWeight == 0 {
		prescription.PatientName = "Миша"
		prescription.PatientGender = "Мужской"
		prescription.PatientWeight = 30.0
		prescription.PatientAge = 5
		r.db.Save(&prescription)
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
		Status:        "черновик",
		DateCreate:    time.Now(),
		CreatorID:     creatorID,
		PatientName:   "Миша",
		PatientGender: "Мужской",
		PatientWeight: 30.0,
		PatientAge:    5,
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

// GetPrescriptionByID - получение заявки по ID (включая черновик, но без удаленных)
func (r *Repository) GetPrescriptionByID(id uint) (*ds.Prescription, error) {
	var prescription ds.Prescription
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("id = ? AND status != ?", id, "удалён").First(&prescription).Error
	if err != nil {
		return nil, err
	}
	return &prescription, nil
}

// GetPrescriptionByIDWithDeleted - получение заявки по ID (включая удаленные, для отображения в истории)
func (r *Repository) GetPrescriptionByIDWithDeleted(id uint) (*ds.Prescription, error) {
	var prescription ds.Prescription
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("id = ?", id).First(&prescription).Error
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
func (r *Repository) AddMedicationToPrescription(prescriptionID, medicationID uint) error {
	// Проверяем, не добавлено ли уже это лекарство
	var existing ds.PrescriptionMedication
	err := r.db.Where("prescription_id = ? AND medication_id = ?", prescriptionID, medicationID).
		First(&existing).Error
	if err == nil {
		// Если уже есть, не добавляем повторно
		return nil
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

// GetAllPrescriptionsForAdmin - получение всех заявок для администратора (включая все статусы)
// Сортировка по дате создания (новые сначала)
func (r *Repository) GetAllPrescriptionsForAdmin() ([]ds.Prescription, error) {
	var prescriptions []ds.Prescription
	err := r.db.Preload("Creator").Preload("Moderator").
		Order("date_create DESC").
		Find(&prescriptions).Error
	return prescriptions, err
}

// GetAllPrescriptionsByUser - получение всех заявок пользователя (включая черновик и удаленные)
// Сортировка по дате создания (новые сначала)
func (r *Repository) GetAllPrescriptionsByUser(creatorID uint) ([]ds.Prescription, error) {
	var prescriptions []ds.Prescription
	err := r.db.Preload("Creator").Preload("Moderator").
		Where("creator_id = ?", creatorID).
		Order("date_create DESC").
		Find(&prescriptions).Error
	return prescriptions, err
}

// GetPrescriptionsWithFilters - получение заявок с фильтрацией
// Если statusFilter пустой, исключает удаленные и черновики
// Если statusFilter указан, фильтрует по нему (включая "удалён" и "черновик")
func (r *Repository) GetPrescriptionsWithFilters(statusFilter string, dateFrom, dateTo *time.Time) ([]ds.Prescription, error) {
	query := r.db.Preload("Creator").Preload("Moderator")

	// Если фильтр по статусу не указан, исключаем удаленные и черновики
	// Если фильтр указан, применяем его (включая "удалён" и "черновик")
	if statusFilter == "" {
		query = query.Where("status != ? AND status != ?", "удалён", "черновик")
	} else {
		// Логируем для отладки
		fmt.Printf("GetPrescriptionsWithFilters: filtering by status='%s'\n", statusFilter)
		query = query.Where("status = ?", statusFilter)
	}

	// Для фильтрации по датам используем date_formation, если она есть
	// Для удаленных записей date_formation может быть NULL, поэтому используем date_create
	if dateFrom != nil || dateTo != nil {
		if statusFilter == "удалён" {
			// Для удаленных записей фильтруем по date_create
			if dateFrom != nil {
				query = query.Where("date_create >= ?", *dateFrom)
			}
			if dateTo != nil {
				query = query.Where("date_create <= ?", *dateTo)
			}
		} else {
			// Для остальных статусов фильтруем по date_formation
			if dateFrom != nil {
				query = query.Where("date_formation >= ?", *dateFrom)
			}
			if dateTo != nil {
				query = query.Where("date_formation <= ?", *dateTo)
			}
		}
	}

	var prescriptions []ds.Prescription
	// Сортировка: для удаленных по date_create, для остальных по date_formation
	if statusFilter == "удалён" {
		err := query.Order("date_create DESC").Find(&prescriptions).Error
		if err != nil {
			return nil, err
		}
		// Логируем для отладки
		fmt.Printf("GetPrescriptionsWithFilters: statusFilter=%s, dateFrom=%v, dateTo=%v, found %d prescriptions\n", 
			statusFilter, dateFrom, dateTo, len(prescriptions))
		return prescriptions, nil
	} else {
		err := query.Order("date_formation DESC").Find(&prescriptions).Error
		return prescriptions, err
	}
}

// GetPrescriptionMedicationsWithDetails - получение лекарств в заявке с деталями
func (r *Repository) GetPrescriptionMedicationsWithDetails(prescriptionID uint) ([]ds.PrescriptionMedication, error) {
	var prescriptionMedications []ds.PrescriptionMedication
	err := r.db.Preload("Medication").
		Where("prescription_id = ?", prescriptionID).
		Order("order_number ASC").
		Find(&prescriptionMedications).Error
	return prescriptionMedications, err
}

// GetPrescriptionMedicationsWithResultCount - количество записей м-м с непустым DosageInstruction
func (r *Repository) GetPrescriptionMedicationsWithResultCount(prescriptionID uint) int64 {
	var count int64
	r.db.Model(&ds.PrescriptionMedication{}).
		Where("prescription_id = ? AND dosage_instruction IS NOT NULL AND dosage_instruction != ''", prescriptionID).
		Count(&count)
	return count
}

// UpdatePrescription - обновление заявки
func (r *Repository) UpdatePrescription(prescription *ds.Prescription) error {
	return r.db.Save(prescription).Error
}

// FormPrescription - формирование заявки создателем
// После формирования заявки черновик остается, но статус меняется на "сформирован"
// Корзина обнуляется автоматически, так как GetDraftPrescription ищет только статус "черновик"
func (r *Repository) FormPrescription(prescriptionID uint) error {
	now := time.Now()
	return r.db.Model(&ds.Prescription{}).
		Where("id = ?", prescriptionID).
		Updates(map[string]interface{}{
			"status":         "сформирован",
			"date_formation": now,
		}).Error
}

// CompleteOrRejectPrescription - завершение/отклонение заявки модератором
func (r *Repository) CompleteOrRejectPrescription(prescriptionID uint, moderatorID uint, newStatus string) error {
	now := time.Now()

	// Получаем заявку для расчета
	prescription, err := r.GetPrescriptionByID(prescriptionID)
	if err != nil {
		return err
	}

	// Получаем лекарства в заявке для расчета
	medications, err := r.GetPrescriptionMedications(prescriptionID)
	if err != nil {
		return err
	}

	// Вычисляем результат (формула из ЛР2)
	calculationResult := ""
	if newStatus == "завершён" {
		calculationResult = r.calculatePrescriptionResult(prescription, medications)
	}

	// Обновляем заявку
	updates := map[string]interface{}{
		"status":       newStatus,
		"moderator_id": moderatorID,
		"date_finish":  now,
	}

	if calculationResult != "" {
		updates["calculation_result"] = calculationResult
	}

	return r.db.Model(&ds.Prescription{}).
		Where("id = ?", prescriptionID).
		Updates(updates).Error
}

// ChangePrescriptionStatus - изменение статуса заявки администратором (любой статус)
func (r *Repository) ChangePrescriptionStatus(prescriptionID uint, moderatorID uint, newStatus string) error {
	now := time.Now()

	// Получаем заявку (включая удаленные)
	prescription, err := r.GetPrescriptionByIDWithDeleted(prescriptionID)
	if err != nil {
		return err
	}

	// Получаем лекарства в заявке для расчета (если статус меняется на "завершён")
	medications, err := r.GetPrescriptionMedications(prescriptionID)
	if err != nil {
		medications = []ds.PrescriptionMedication{} // Если ошибка, продолжаем без расчета
	}

	// Вычисляем результат только если статус меняется на "завершён"
	calculationResult := ""
	if newStatus == "завершён" && len(medications) > 0 {
		calculationResult = r.calculatePrescriptionResult(prescription, medications)
	}

	// Обновляем заявку
	updates := map[string]interface{}{
		"status": newStatus,
	}

	// Устанавливаем модератора только для статусов, которые требуют модератора
	if newStatus == "завершён" || newStatus == "отклонён" || newStatus == "удалён" {
		updates["moderator_id"] = moderatorID
	} else {
		// Для других статусов обнуляем модератора
		updates["moderator_id"] = nil
	}

	// Устанавливаем дату завершения для статусов завершён/отклонён
	if newStatus == "завершён" || newStatus == "отклонён" {
		updates["date_finish"] = now
	} else {
		// Для других статусов обнуляем дату завершения
		updates["date_finish"] = nil
	}

	// Устанавливаем дату формирования для статуса "сформирован"
	if newStatus == "сформирован" && !prescription.DateFormation.Valid {
		updates["date_formation"] = now
	}

	// Устанавливаем результат расчета только для статуса "завершён"
	if newStatus == "завершён" && calculationResult != "" {
		updates["calculation_result"] = calculationResult
	} else if newStatus != "завершён" {
		// Для других статусов очищаем результат расчета
		updates["calculation_result"] = nil
	}

	// Выполняем обновление
	result := r.db.Model(&ds.Prescription{}).
		Where("id = ?", prescriptionID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	// Проверяем, что обновление прошло успешно
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// calculatePrescriptionResult - расчет результата заявки
// Использует две формулы и берет среднее арифметическое:
// 1. Доза для ребёнка = (Доза взрослого × масса ребёнка в кг) / 70
// 2. Доза для ребёнка = (Доза взрослого × число лет ребёнка) / 24
// Результат = (формула1 + формула2) / 2
func (r *Repository) calculatePrescriptionResult(prescription *ds.Prescription, medications []ds.PrescriptionMedication) string {
	if len(medications) == 0 {
		return ""
	}

	if prescription.PatientWeight <= 0 || prescription.PatientAge <= 0 {
		return "Ошибка: не указаны вес или возраст пациента"
	}

	result := "Расчет дозы:\n"
	for _, pm := range medications {
		if pm.Medication.Name != "" {
			// Получаем дозу взрослого из таблицы препаратов
			adultDose := pm.Medication.AdultDose
			if adultDose <= 0 {
				// Если доза не указана, пропускаем этот препарат
				result += fmt.Sprintf("- %s: доза для взрослого не указана, расчет невозможен\n", pm.Medication.Name)
				continue
			}

			// Формула 1: (Доза взрослого × масса ребёнка в кг) / 70
			doseByWeight := (adultDose * prescription.PatientWeight) / 70.0

			// Формула 2: (Доза взрослого × число лет ребёнка) / 24
			doseByAge := (adultDose * float64(prescription.PatientAge)) / 24.0

			// Среднее арифметическое
			finalDose := (doseByWeight + doseByAge) / 2.0

			// Минимальная доза
			if finalDose < 0.1 {
				finalDose = 0.1
			}

			result += fmt.Sprintf("- %s:\n", pm.Medication.Name)
			result += fmt.Sprintf("  Доза взрослого: %.2f мг\n", adultDose)
			result += fmt.Sprintf("  По весу (%.2f кг): %.2f мг\n", prescription.PatientWeight, doseByWeight)
			result += fmt.Sprintf("  По возрасту (%d лет): %.2f мг\n", prescription.PatientAge, doseByAge)
			result += fmt.Sprintf("  Итоговая доза (среднее): %.2f мг\n", finalDose)
		}
	}

	result += fmt.Sprintf("\nИтог: расчет дозы произведен для пациента %s (возраст: %d лет, вес: %.2f кг)",
		prescription.PatientName, prescription.PatientAge, prescription.PatientWeight)

	return result
}
