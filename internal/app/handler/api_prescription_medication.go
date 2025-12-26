package handler

import (
	"metoda/internal/app/auth"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// DeletePrescriptionMedicationAPI - DELETE /api/prescriptions/:id/medications/:medication_id
func (h *Handler) DeletePrescriptionMedicationAPI(ctx *gin.Context) {
	prescriptionIDStr := ctx.Param("id")
	medicationIDStr := ctx.Param("medication_id")

	prescriptionID, err := strconv.ParseUint(prescriptionIDStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid prescription ID format")
		return
	}

	medicationID, err := strconv.ParseUint(medicationIDStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid medication ID format")
		return
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	// Проверяем, что заявка принадлежит создателю
	prescription, err := h.Repository.GetPrescriptionByID(uint(prescriptionID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if prescription.CreatorID != creatorID {
		h.errorResponse(ctx, http.StatusForbidden, "You can only modify your own prescriptions")
		return
	}

	// Проверяем, что заявка в статусе черновик
	if prescription.Status != "черновик" {
		h.errorResponse(ctx, http.StatusBadRequest, "Can only delete medications from draft prescriptions")
		return
	}

	// Удаляем запись м-м
	err = h.Repository.DeletePrescriptionMedication(uint(prescriptionID), uint(medicationID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Medication not found in prescription")
			return
		}
		logrus.Errorf("Error deleting prescription medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to delete medication from prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Medication removed from prescription",
	})
}

// UpdatePrescriptionMedicationAPI - PUT /api/prescriptions/:id/medications/:medication_id
func (h *Handler) UpdatePrescriptionMedicationAPI(ctx *gin.Context) {
	prescriptionIDStr := ctx.Param("id")
	medicationIDStr := ctx.Param("medication_id")

	prescriptionID, err := strconv.ParseUint(prescriptionIDStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid prescription ID format")
		return
	}

	medicationID, err := strconv.ParseUint(medicationIDStr, 10, 32)
	if err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid medication ID format")
		return
	}

	var req struct {
		OrderNumber       *int    `json:"order_number"`
		DosageInstruction *string `json:"dosage_instruction"`
		IsMain            *bool   `json:"is_main"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// Получаем ID фиксированного создателя через singleton
	creatorID := auth.GetUserService().GetCreatorID()

	// Проверяем, что заявка принадлежит создателю
	prescription, err := h.Repository.GetPrescriptionByID(uint(prescriptionID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Prescription not found")
			return
		}
		logrus.Errorf("Error getting prescription: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if prescription.CreatorID != creatorID {
		h.errorResponse(ctx, http.StatusForbidden, "You can only modify your own prescriptions")
		return
	}

	// Проверяем, что заявка в статусе черновик
	if prescription.Status != "черновик" {
		h.errorResponse(ctx, http.StatusBadRequest, "Can only update medications in draft prescriptions")
		return
	}

	// Обновляем запись м-м
	err = h.Repository.UpdatePrescriptionMedication(uint(prescriptionID), uint(medicationID), req.OrderNumber, req.DosageInstruction, req.IsMain)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "Medication not found in prescription")
			return
		}
		logrus.Errorf("Error updating prescription medication: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to update medication in prescription")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Medication updated in prescription",
	})
}
