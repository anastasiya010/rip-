package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"metoda/internal/app/auth"
	"metoda/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// GetCartIconAPI - GET /api/prescriptions/cart-icon - иконка корзины
func (h *Handler) GetCartIconAPI(ctx *gin.Context) {
	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	// Получаем ID черновика заявки
	draftID := h.Repository.GetDraftPrescriptionID(creatorID)
	if draftID == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": map[string]interface{}{
				"prescription_id": nil,
				"count":           0,
			},
		})
		return
	}

	// Получаем количество услуг в заявке
	count := h.Repository.GetCartCount(creatorID)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"prescription_id": *draftID,
			"count":           count,
		},
	})
}

// GetPrescriptionsAPI - GET /api/prescriptions - список заявок с фильтрацией
func (h *Handler) GetPrescriptionsAPI(ctx *gin.Context) {
	// Фильтрация по статусу
	statusFilter := strings.TrimSpace(ctx.Query("status"))

	// Фильтрация по диапазону даты формирования
	dateFromStr := strings.TrimSpace(ctx.Query("date_from"))
	dateToStr := strings.TrimSpace(ctx.Query("date_to"))

	var dateFrom, dateTo *time.Time
	if dateFromStr != "" {
		parsed, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			h.errorResponse(ctx, http.StatusBadRequest, "Invalid date_from format. Use YYYY-MM-DD")
			return
		}
		// Используем UTC для консистентности с базой данных
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		dateFrom = &parsed
	}
	if dateToStr != "" {
		parsed, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			h.errorResponse(ctx, http.StatusBadRequest, "Invalid date_to format. Use YYYY-MM-DD")
			return
		}
		// Устанавливаем время на конец дня в UTC
		parsed = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 23, 59, 59, 999999999, time.UTC)
		dateTo = &parsed
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()
	isAdmin := auth.IsModerator(ctx)

	// Получаем заявки с фильтрацией
	prescriptions, err := h.Repository.GetPrescriptionsWithFilters(statusFilter, dateFrom, dateTo)
	if err != nil {
		logrus.Errorf("Error getting prescriptions: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Логируем для отладки
	logrus.Infof("GetPrescriptionsAPI: statusFilter='%s', isAdmin=%v, creatorID=%d, found %d prescriptions before filtering", statusFilter, isAdmin, creatorID, len(prescriptions))
	if len(prescriptions) > 0 {
		for i, p := range prescriptions {
			logrus.Infof("  Prescription[%d]: ID=%d, Status=%s, CreatorID=%d", i, p.ID, p.Status, p.CreatorID)
		}
	}

	// Если пользователь не администратор, фильтруем только заявки фиксированного создателя
	if !isAdmin {
		filtered := []ds.Prescription{}
		for _, presc := range prescriptions {
			if presc.CreatorID == creatorID {
				filtered = append(filtered, presc)
			}
		}
		prescriptions = filtered
		logrus.Infof("GetPrescriptionsAPI: after filtering by creator_id=%d, found %d prescriptions", creatorID, len(prescriptions))
	}

	// Преобразуем в формат ответа
	response := make([]map[string]interface{}, len(prescriptions))
	for i, presc := range prescriptions {
		// Вычисляемое поле: количество записей м-м с непустым DosageInstruction
		calculatedCount := h.Repository.GetPrescriptionMedicationsWithResultCount(presc.ID)

		creatorLogin := ""
		if presc.Creator.Login != "" {
			creatorLogin = presc.Creator.Login
		}

		moderatorLogin := ""
		if presc.Moderator != nil && presc.Moderator.Login != "" {
			moderatorLogin = presc.Moderator.Login
		}

		dateFormation := ""
		if presc.DateFormation.Valid {
			dateFormation = presc.DateFormation.Time.Format(time.RFC3339)
		}

		dateFinish := ""
		if presc.DateFinish.Valid {
			dateFinish = presc.DateFinish.Time.Format(time.RFC3339)
		}

		response[i] = map[string]interface{}{
			"id":                 presc.ID,
			"status":             presc.Status,
			"date_create":        presc.DateCreate.Format(time.RFC3339),
			"date_formation":     dateFormation,
			"date_finish":        dateFinish,
			"creator_login":      creatorLogin,
			"moderator_login":    moderatorLogin,
			"patient_name":       presc.PatientName,
			"patient_gender":     presc.PatientGender,
			"patient_weight":     presc.PatientWeight,
			"patient_age":        presc.PatientAge,
			"calculation_result": presc.CalculationResult,
			"medications_count":  calculatedCount, // Вычисляемое поле
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetPrescriptionByIDAPI - GET /api/prescriptions/:id - одна запись заявки
func (h *Handler) GetPrescriptionByIDAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	prescription, err := h.Repository.GetPrescriptionByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Получаем лекарства в заявке
	medications, err := h.Repository.GetPrescriptionMedicationsWithDetails(uint(id))
	if err != nil {
		logrus.Errorf("Error getting prescription medications: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Преобразуем лекарства в формат ответа с изображениями
	medicationsResponse := make([]map[string]interface{}, len(medications))
	for i, med := range medications {
		imageURL := ""
		if med.Medication.ImageURL != nil && *med.Medication.ImageURL != "" {
			imageURL = h.Minio.GetImageURL(*med.Medication.ImageURL)
		}

		medicationsResponse[i] = map[string]interface{}{
			"medication_id":      med.MedicationID,
			"medication_name":    med.Medication.Name,
			"image_url":          imageURL,
			"order_number":       med.OrderNumber,
			"is_main":            med.IsMain,
			"dosage_instruction": med.DosageInstruction,
		}
	}

	creatorLogin := ""
	if prescription.Creator.Login != "" {
		creatorLogin = prescription.Creator.Login
	}

	moderatorLogin := ""
	if prescription.Moderator != nil && prescription.Moderator.Login != "" {
		moderatorLogin = prescription.Moderator.Login
	}

	dateFormation := ""
	if prescription.DateFormation.Valid {
		dateFormation = prescription.DateFormation.Time.Format(time.RFC3339)
	}

	dateFinish := ""
	if prescription.DateFinish.Valid {
		dateFinish = prescription.DateFinish.Time.Format(time.RFC3339)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":                 prescription.ID,
			"status":             prescription.Status,
			"date_create":        prescription.DateCreate.Format(time.RFC3339),
			"date_formation":     dateFormation,
			"date_finish":        dateFinish,
			"creator_login":      creatorLogin,
			"moderator_login":    moderatorLogin,
			"patient_name":       prescription.PatientName,
			"patient_gender":     prescription.PatientGender,
			"patient_weight":     prescription.PatientWeight,
			"patient_age":        prescription.PatientAge,
			"calculation_result": prescription.CalculationResult,
			"medications":        medicationsResponse,
		},
	})
}

// UpdatePrescriptionAPI - PUT /api/prescriptions/:id - изменения полей заявки по теме
func (h *Handler) UpdatePrescriptionAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req struct {
		PatientName   string  `json:"patient_name"`
		PatientGender string  `json:"patient_gender"`
		PatientWeight float64 `json:"patient_weight"`
		PatientAge    int     `json:"patient_age"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	prescription, err := h.Repository.GetPrescriptionByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Проверяем, что заявка не удалена
	if prescription.Status == "удалён" {
		h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
		return
	}

	// Обновляем только переданные поля
	if req.PatientName != "" {
		prescription.PatientName = req.PatientName
	}
	if req.PatientGender != "" {
		prescription.PatientGender = req.PatientGender
	}
	if req.PatientWeight > 0 {
		prescription.PatientWeight = req.PatientWeight
	}
	if req.PatientAge > 0 {
		prescription.PatientAge = req.PatientAge
	}

	err = h.Repository.UpdatePrescription(prescription)
	if err != nil {
		logrus.Errorf("Error updating prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to update prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Prescription updated successfully",
	})
}

// FormPrescriptionAPI - PUT /api/prescriptions/:id/form - сформировать создателем
func (h *Handler) FormPrescriptionAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	prescription, err := h.Repository.GetPrescriptionByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Проверяем права доступа
	if prescription.CreatorID != creatorID {
		h.errorResponse(ctx, http.StatusForbidden, "You can only form your own prescriptions")
		return
	}

	// Проверяем статус
	if prescription.Status != "черновик" {
		h.errorResponse(ctx, http.StatusBadRequest, "Only draft prescriptions can be formed")
		return
	}

	// Проверка на обязательные поля
	if prescription.PatientName == "" || prescription.PatientAge == 0 || prescription.PatientWeight == 0 {
		h.errorResponse(ctx, http.StatusBadRequest, "Missing required fields: patient_name, patient_age, patient_weight")
		return
	}

	// Проверяем, что в заявке есть лекарства
	medications, err := h.Repository.GetPrescriptionMedications(prescription.ID)
	if err != nil {
		logrus.Errorf("Error getting medications: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if len(medications) == 0 {
		h.errorResponse(ctx, http.StatusBadRequest, "Prescription must contain at least one medication")
		return
	}

	// Обновляем статус и дату формирования
	err = h.Repository.FormPrescription(uint(id))
	if err != nil {
		logrus.Errorf("Error forming prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to form prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Prescription formed successfully",
	})
}

// CompleteOrRejectPrescriptionAPI - PUT /api/prescriptions/:id/complete или /api/prescriptions/:id/reject
func (h *Handler) CompleteOrRejectPrescriptionAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Определяем действие из полного пути
	fullPath := ctx.FullPath()
	var action string
	if strings.Contains(fullPath, "/complete") {
		action = "complete"
	} else if strings.Contains(fullPath, "/reject") {
		action = "reject"
	} else {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid action. Use 'complete' or 'reject'")
		return
	}

	// Проверяем, что пользователь является администратором
	isModerator := auth.IsModerator(ctx)
	userID, userIDExists := auth.GetUserID(ctx)
	userLogin, loginExists := auth.GetUserLogin(ctx)
	
	// Детальная диагностика
	logrus.Infof("CompleteOrRejectPrescriptionAPI: isModerator=%v, userID=%v (exists=%v), userLogin=%v (exists=%v)", 
		isModerator, userID, userIDExists, userLogin, loginExists)
	
	// Проверяем значение в контексте напрямую
	if isModValue, exists := ctx.Get("is_moderator"); exists {
		logrus.Infof("CompleteOrRejectPrescriptionAPI: Direct context check - is_moderator=%v (type=%T, exists=%v)", 
			isModValue, isModValue, exists)
	}
	
	if !isModerator {
		logrus.Warnf("CompleteOrRejectPrescriptionAPI: Access denied - UserID=%v, Login=%v, IsModerator=%v", 
			userID, userLogin, isModerator)
		h.errorResponse(ctx, http.StatusForbidden, "Only administrators can complete or reject prescriptions")
		return
	}

	// Получаем ID модератора из сессии
	moderatorID, exists := auth.GetUserID(ctx)
	if !exists {
		h.errorResponse(ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	prescription, err := h.Repository.GetPrescriptionByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Проверяем статус
	if prescription.Status != "сформирован" {
		h.errorResponse(ctx, http.StatusBadRequest, "Only formed prescriptions can be completed or rejected")
		return
	}

	newStatus := "завершён"
	if action == "reject" {
		newStatus = "отклонён"
	}

	// Завершаем/отклоняем заявку
	err = h.Repository.CompleteOrRejectPrescription(uint(id), moderatorID, newStatus)
	if err != nil {
		logrus.Errorf("Error completing/rejecting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to process prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Prescription " + action + "ed successfully",
	})
}

// DeletePrescriptionAPI - DELETE /api/prescriptions/:id - удаление заявки
func (h *Handler) DeletePrescriptionAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	prescription, err := h.Repository.GetPrescriptionByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Проверяем права доступа
	if prescription.CreatorID != creatorID {
		h.errorResponse(ctx, http.StatusForbidden, "You can only delete your own prescriptions")
		return
	}

	// Проверяем статус - можно удалять только черновики
	if prescription.Status != "черновик" {
		h.errorResponse(ctx, http.StatusBadRequest, "Only draft prescriptions can be deleted")
		return
	}

	// Логическое удаление
	err = h.Repository.DeletePrescription(uint(id))
	if err != nil {
		logrus.Errorf("Error deleting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to delete prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Prescription deleted successfully",
	})
}
