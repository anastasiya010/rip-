package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"metoda/internal/app/ds"
)

// GetPrescription - GET метод для просмотра текущей заявки (черновик)
func (h *Handler) GetPrescription(ctx *gin.Context) {
	creatorID := uint(1) // Временно захардкожено, в будущем из JWT

	prescription, err := h.Repository.GetDraftPrescription(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if prescription == nil {
		ctx.HTML(http.StatusOK, "layout", gin.H{
			"TemplateName":                "order",
			"Prescription":                nil,
			"PrescriptionMedicationCount": 0,
			"ImageURL": func(key string) string {
				if key == "" || key == "null" {
					return "/static/img/defaultIcon.png"
				}
				// Если путь уже полный URL (от MinIO), возвращаем как есть
				if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
					return key
				}
				// Если путь уже содержит /static/img/, возвращаем как есть
				if strings.HasPrefix(key, "/static/img/") {
					return key
				}
				// Получаем базовый URL из переменной окружения или используем /static/img
				baseURL := os.Getenv("MINIO_PUBLIC_BASE")
				if baseURL != "" {
					baseURL = strings.TrimRight(baseURL, "/")
					return baseURL + "/" + key
				}
				// По умолчанию используем локальный путь
				return "/static/img/" + key
			},
		})
		return
	}

	// Получаем лекарства в заявке
	prescriptionMedications, err := h.Repository.GetPrescriptionMedications(prescription.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Преобразуем в формат для шаблона
	var medications []struct {
		MedicationCardID  uint
		MedicationName    string
		ImageKey          string
		Quantity          int
		DosageInstruction string
	}

	for _, pm := range prescriptionMedications {
		imageKey := ""
		if pm.Medication.ImageURL != nil {
			imageKey = *pm.Medication.ImageURL
		}
		medications = append(medications, struct {
			MedicationCardID  uint
			MedicationName    string
			ImageKey          string
			Quantity          int
			DosageInstruction string
		}{
			MedicationCardID:  pm.Medication.ID,
			MedicationName:    pm.Medication.Name,
			ImageKey:          imageKey,
			Quantity:          pm.Quantity,
			DosageInstruction: pm.DosageInstruction,
		})
	}

	prescriptionData := &ds.Prescription{
		ID:                  prescription.ID,
		Status:              prescription.Status,
		PatientName:          prescription.PatientName,
		PatientGender:        prescription.PatientGender,
		PatientWeight:        prescription.PatientWeight,
		PatientAge:           prescription.PatientAge,
		CalculationResult:   prescription.CalculationResult,
	}

	ctx.HTML(http.StatusOK, "layout", gin.H{
		"TemplateName": "order",
		"Prescription": struct {
			ID                uint
			Status            string
			PatientName       string
			PatientGender     string
			PatientWeight     float64
			PatientAge        int
			CalculationResult string
			Medications       []struct {
				MedicationCardID  uint
				MedicationName    string
				ImageKey          string
				Quantity          int
				DosageInstruction string
			}
		}{
			ID:                prescriptionData.ID,
			Status:            prescriptionData.Status,
			PatientName:       prescriptionData.PatientName,
			PatientGender:     prescriptionData.PatientGender,
			PatientWeight:     prescriptionData.PatientWeight,
			PatientAge:        prescriptionData.PatientAge,
			CalculationResult: prescriptionData.CalculationResult,
			Medications:       medications,
		},
		"PrescriptionMedicationCount": len(medications),
		"ImageURL": func(key string) string {
			if key == "" || key == "null" {
				return "/static/img/defaultIcon.png"
			}
			// Если путь уже полный URL (от MinIO), возвращаем как есть
			if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
				return key
			}
			// Если путь уже содержит /static/img/, возвращаем как есть
			if strings.HasPrefix(key, "/static/img/") {
				return key
			}
			// Получаем базовый URL из переменной окружения или используем /static/img
			baseURL := os.Getenv("MINIO_PUBLIC_BASE")
			if baseURL != "" {
				baseURL = strings.TrimRight(baseURL, "/")
				return baseURL + "/" + key
			}
			// По умолчанию используем локальный путь
			return "/static/img/" + key
		},
	})
}

// AddMedicationToPrescription - POST метод для добавления лекарства в заявку через ORM
func (h *Handler) AddMedicationToPrescription(ctx *gin.Context) {
	creatorID := uint(1) // Временно захардкожено, в будущем из JWT

	medicationIDStr := ctx.PostForm("medication_id")
	medicationID, err := strconv.ParseUint(medicationIDStr, 10, 32)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid medication ID"})
		return
	}

	quantityStr := ctx.PostForm("quantity")
	quantity := 1
	if quantityStr != "" {
		q, err := strconv.Atoi(quantityStr)
		if err == nil && q > 0 {
			quantity = q
		}
	}

	// Получаем или создаем черновик заявки
	prescription, err := h.Repository.GetDraftPrescription(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	if prescription == nil {
		prescription, err = h.Repository.CreateDraftPrescription(creatorID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}
	}

	// Добавляем лекарство в заявку
	err = h.Repository.AddMedicationToPrescription(prescription.ID, uint(medicationID), quantity)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/catalog")
}

// DeletePrescription - POST метод для логического удаления заявки через SQL UPDATE
func (h *Handler) DeletePrescription(ctx *gin.Context) {
	prescriptionIDStr := ctx.PostForm("prescription_id")
	prescriptionID, err := strconv.ParseUint(prescriptionIDStr, 10, 32)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prescription ID"})
		return
	}

	err = h.Repository.DeletePrescription(uint(prescriptionID))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Redirect(http.StatusFound, "/catalog")
}

