package handler

import (
	"net/http"
	"strconv"
	"strings"

	"metoda/internal/app/auth"
	"metoda/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetMedicationsAPI - GET /api/medications - список услуг с фильтрацией
func (h *Handler) GetMedicationsAPI(ctx *gin.Context) {
	// Фильтрация по категории и производителю
	categoryFilter := strings.TrimSpace(ctx.Query("category"))
	manufacturerFilter := strings.TrimSpace(ctx.Query("manufacturer"))

	// Фильтрация на бэкенде через SQL запрос
	medications, err := h.Repository.GetMedicationsWithFilters(categoryFilter, manufacturerFilter)
	if err != nil {
		logrus.Errorf("Error getting medications: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Преобразуем ImageURL в полный URL
	response := make([]map[string]interface{}, len(medications))
	for i, med := range medications {
		imageURL := ""
		if med.ImageURL != nil && *med.ImageURL != "" {
			imageURL = h.Minio.GetImageURL(*med.ImageURL)
		}

		response[i] = map[string]interface{}{
			"id":           med.ID,
			"name":         med.Name,
			"description":  med.Description,
			"category":     med.Category,
			"manufacturer": med.Manufacturer,
			"short_info":   med.ShortInfo,
			"image_url":    imageURL,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   response,
	})
}

// GetMedicationByIDAPI - GET /api/medications/:id - одна запись услуги
func (h *Handler) GetMedicationByIDAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	medication, err := h.Repository.GetMedicationByID(uint(id))
	if err != nil {
		logrus.Errorf("Error getting medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if medication == nil || medication.IsDeleted {
		h.errorResponse(ctx, http.StatusNotFound, "Medication not found")
		return
	}

	imageURL := ""
	if medication.ImageURL != nil && *medication.ImageURL != "" {
		imageURL = h.Minio.GetImageURL(*medication.ImageURL)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           medication.ID,
			"name":         medication.Name,
			"description":  medication.Description,
			"category":     medication.Category,
			"manufacturer": medication.Manufacturer,
			"short_info":   medication.ShortInfo,
			"image_url":    imageURL,
		},
	})
}

// CreateMedicationAPI - POST /api/medications - добавление услуги (без изображения)
func (h *Handler) CreateMedicationAPI(ctx *gin.Context) {
	var req struct {
		Name         string `json:"name" binding:"required"`
		Description  string `json:"description"`
		Category     string `json:"category"`
		Manufacturer string `json:"manufacturer"`
		ShortInfo    string `json:"short_info"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	medication := &ds.Medication{
		Name:         req.Name,
		Description:  req.Description,
		Category:     req.Category,
		Manufacturer: req.Manufacturer,
		ShortInfo:    req.ShortInfo,
		IsDeleted:    false,
	}

	err := h.Repository.CreateMedication(medication)
	if err != nil {
		logrus.Errorf("Error creating medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to create medication")
		return
	}

	imageURL := ""
	if medication.ImageURL != nil && *medication.ImageURL != "" {
		imageURL = h.Minio.GetImageURL(*medication.ImageURL)
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           medication.ID,
			"name":         medication.Name,
			"description":  medication.Description,
			"category":     medication.Category,
			"manufacturer": medication.Manufacturer,
			"short_info":   medication.ShortInfo,
			"image_url":    imageURL,
		},
	})
}

// UpdateMedicationAPI - PUT /api/medications/:id - изменение услуги
func (h *Handler) UpdateMedicationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req struct {
		Name         string `json:"name"`
		Description  string `json:"description"`
		Category     string `json:"category"`
		Manufacturer string `json:"manufacturer"`
		ShortInfo    string `json:"short_info"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	medication, err := h.Repository.GetMedicationByID(uint(id))
	if err != nil {
		logrus.Errorf("Error getting medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if medication == nil || medication.IsDeleted {
		h.errorResponse(ctx, http.StatusNotFound, "Medication not found")
		return
	}

	// Обновляем только переданные поля
	if req.Name != "" {
		medication.Name = req.Name
	}
	if req.Description != "" {
		medication.Description = req.Description
	}
	if req.Category != "" {
		medication.Category = req.Category
	}
	if req.Manufacturer != "" {
		medication.Manufacturer = req.Manufacturer
	}
	if req.ShortInfo != "" {
		medication.ShortInfo = req.ShortInfo
	}

	err = h.Repository.UpdateMedication(medication)
	if err != nil {
		logrus.Errorf("Error updating medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to update medication")
		return
	}

	imageURL := ""
	if medication.ImageURL != nil && *medication.ImageURL != "" {
		imageURL = h.Minio.GetImageURL(*medication.ImageURL)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           medication.ID,
			"name":         medication.Name,
			"description":  medication.Description,
			"category":     medication.Category,
			"manufacturer": medication.Manufacturer,
			"short_info":   medication.ShortInfo,
			"image_url":    imageURL,
		},
	})
}

// DeleteMedicationAPI - DELETE /api/medications/:id - удаление услуги
func (h *Handler) DeleteMedicationAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	medication, err := h.Repository.GetMedicationByID(uint(id))
	if err != nil {
		logrus.Errorf("Error getting medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if medication == nil || medication.IsDeleted {
		h.errorResponse(ctx, http.StatusNotFound, "Medication not found")
		return
	}

	// Удаляем изображение из Minio если есть
	if medication.ImageURL != nil && *medication.ImageURL != "" {
		err = h.Minio.DeleteImage(ctx.Request.Context(), *medication.ImageURL)
		if err != nil {
			logrus.Warnf("Failed to delete image from Minio: %v", err)
		}
	}

	// Логическое удаление
	err = h.Repository.DeleteMedication(uint(id))
	if err != nil {
		logrus.Errorf("Error deleting medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to delete medication")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Medication deleted successfully",
	})
}

// AddMedicationToDraftAPI - POST /api/medications/:id/add-to-draft - добавление в заявку-черновик
func (h *Handler) AddMedicationToDraftAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	medicationID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid medication ID format")
		return
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	// Получаем или создаем черновик заявки
	prescription, err := h.Repository.GetDraftPrescription(creatorID)
	if err != nil {
		logrus.Errorf("Error getting draft prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if prescription == nil {
		prescription, err = h.Repository.CreateDraftPrescription(creatorID)
		if err != nil {
			logrus.Errorf("Error creating draft prescription: %v", err)
			h.errorResponse(ctx, http.StatusInternalServerError, "Failed to create draft prescription")
			return
		}
	}

	// Добавляем лекарство в заявку
	err = h.Repository.AddMedicationToPrescription(prescription.ID, uint(medicationID))
	if err != nil {
		logrus.Errorf("Error adding medication to prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to add medication to prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Medication added to draft prescription",
		"data": map[string]interface{}{
			"prescription_id": prescription.ID,
		},
	})
}

// UploadMedicationImageAPI - POST /api/medications/:id/image - добавление изображения
func (h *Handler) UploadMedicationImageAPI(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid ID format")
		return
	}

	medication, err := h.Repository.GetMedicationByID(uint(id))
	if err != nil {
		logrus.Errorf("Error getting medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if medication == nil || medication.IsDeleted {
		h.errorResponse(ctx, http.StatusNotFound, "Medication not found")
		return
	}

	// Получаем файл из формы
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "No image file provided")
		return
	}

	// Открываем файл
	src, err := file.Open()
	if err != nil {
		logrus.Errorf("Error opening file: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to process image")
		return
	}
	defer src.Close()

	// Удаляем старое изображение если есть
	if medication.ImageURL != nil && *medication.ImageURL != "" {
		err = h.Minio.DeleteImage(ctx.Request.Context(), *medication.ImageURL)
		if err != nil {
			logrus.Warnf("Failed to delete old image: %v", err)
		}
	}

	// Загружаем новое изображение
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	objectName, err := h.Minio.UploadImage(ctx.Request.Context(), src, contentType)
	if err != nil {
		logrus.Errorf("Error uploading image: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to upload image")
		return
	}

	// Обновляем запись в БД
	medication.ImageURL = &objectName
	err = h.Repository.UpdateMedication(medication)
	if err != nil {
		logrus.Errorf("Error updating medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to update medication")
		return
	}

	imageURL := h.Minio.GetImageURL(objectName)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"image_url": imageURL,
		},
	})
}

// errorResponse - вспомогательная функция для ответов с ошибками
func (h *Handler) errorResponse(ctx *gin.Context, statusCode int, message string) {
	// Явно устанавливаем Content-Type для JSON
	ctx.Header("Content-Type", "application/json; charset=utf-8")
	ctx.JSON(statusCode, gin.H{
		"status":  "fail",
		"message": message,
	})
}
