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

// GetMedications - GET метод для получения и поиска лекарств
func (h *Handler) GetMedications(ctx *gin.Context) {
	var medications []ds.Medication
	var err error

	searchQuery := ctx.Query("search")
	creatorID := uint(1) // Временно захардкожено, в будущем из JWT

	if searchQuery == "" {
		medications, err = h.Repository.GetAllMedications()
	} else {
		medications, err = h.Repository.SearchMedicationsByName(searchQuery)
	}

	if err != nil {
		logrus.Errorf("Error getting medications: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Если нет лекарств, все равно показываем страницу
	if medications == nil {
		medications = []ds.Medication{}
	}

	// Получаем количество лекарств в корзине
	cartCount := h.Repository.GetCartCount(creatorID)
	draftID := h.Repository.GetDraftPrescriptionID(creatorID)
	
	// Инициализируем пустые слайсы, если они nil
	if medications == nil {
		medications = []ds.Medication{}
	}

	// Собираем уникальные категории и производители
	categoryMap := make(map[string]struct{})
	manufacturerMap := make(map[string]struct{})
	for _, med := range medications {
		if med.Category != "" {
			categoryMap[med.Category] = struct{}{}
		}
		if med.Manufacturer != "" {
			manufacturerMap[med.Manufacturer] = struct{}{}
		}
	}

	var categories, manufacturers []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	for man := range manufacturerMap {
		manufacturers = append(manufacturers, man)
	}
	
	// Инициализируем пустые слайсы, если они nil
	if categories == nil {
		categories = []string{}
	}
	if manufacturers == nil {
		manufacturers = []string{}
	}

	selectedCategory := ctx.Query("cat")
	selectedManufacturer := ctx.Query("man")

	// Фильтрация по категории и производителю
	if selectedCategory != "" || selectedManufacturer != "" {
		filtered := []ds.Medication{}
		for _, med := range medications {
			if selectedCategory != "" && !strings.EqualFold(med.Category, selectedCategory) {
				continue
			}
			if selectedManufacturer != "" && !strings.EqualFold(med.Manufacturer, selectedManufacturer) {
				continue
			}
			filtered = append(filtered, med)
		}
		medications = filtered
	}

	var draftIDValue interface{}
	var hasDraft bool
	if draftID != nil {
		draftIDValue = *draftID
		hasDraft = true
	} else {
		draftIDValue = nil
		hasDraft = false
	}

	data := gin.H{
		"TemplateName":                 "catalog",
		"MedicationCards":             medications,
		"PrescriptionMedicationCount": cartCount,
		"cart_count":                  cartCount,
		"draft_id":                    draftIDValue,
		"DraftID":                     draftIDValue,
		"HasDraft":                    hasDraft,
		"Search":                      searchQuery,
		"SelectedCategory":            selectedCategory,
		"SelectedManufacturer":        selectedManufacturer,
		"Categories":                  categories,
		"Manufacturers":               manufacturers,
		"ImageURL": func(key interface{}) string {
			if key == nil {
				return "/static/img/defaultIcon.png"
			}
			var imgPath string
			if ptr, ok := key.(*string); ok && ptr != nil {
				imgPath = *ptr
			} else if str, ok := key.(string); ok {
				imgPath = str
			}
			if imgPath == "" {
				return "/static/img/defaultIcon.png"
			}
			// Если путь уже полный URL (от MinIO), возвращаем как есть
			if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
				return imgPath
			}
			// Если путь уже содержит /static/img/, возвращаем как есть
			if strings.HasPrefix(imgPath, "/static/img/") {
				return imgPath
			}
			// Получаем базовый URL из переменной окружения или используем /static/img
			baseURL := os.Getenv("MINIO_PUBLIC_BASE")
			if baseURL != "" {
				baseURL = strings.TrimRight(baseURL, "/")
				return baseURL + "/" + imgPath
			}
			// По умолчанию используем локальный путь
			return "/static/img/" + imgPath
		},
	}
	
	logrus.Infof("Rendering catalog page with %d medications, TemplateName: %s", len(medications), data["TemplateName"])
	ctx.HTML(http.StatusOK, "layout", data)
}

// GetMedicationByID - GET метод для получения детальной информации о лекарстве
func (h *Handler) GetMedicationByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	medication, err := h.Repository.GetMedicationByID(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Medication not found"})
		return
	}

	if medication == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Medication not found"})
		return
	}

	ctx.HTML(http.StatusOK, "layout", gin.H{
		"TemplateName":  "service",
		"MedicationCard": medication,
		"ImageURL": func(key interface{}) string {
			if key == nil {
				return "/static/img/defaultIcon.png"
			}
			var imgPath string
			if ptr, ok := key.(*string); ok && ptr != nil {
				imgPath = *ptr
			} else if str, ok := key.(string); ok {
				imgPath = str
			}
			if imgPath == "" {
				return "/static/img/defaultIcon.png"
			}
			// Если путь уже полный URL (от MinIO), возвращаем как есть
			if strings.HasPrefix(imgPath, "http://") || strings.HasPrefix(imgPath, "https://") {
				return imgPath
			}
			// Если путь уже содержит /static/img/, возвращаем как есть
			if strings.HasPrefix(imgPath, "/static/img/") {
				return imgPath
			}
			// Получаем базовый URL из переменной окружения или используем /static/img
			baseURL := os.Getenv("MINIO_PUBLIC_BASE")
			if baseURL != "" {
				baseURL = strings.TrimRight(baseURL, "/")
				return baseURL + "/" + imgPath
			}
			// По умолчанию используем локальный путь
			return "/static/img/" + imgPath
		},
	})
}

// DeleteMedication - POST метод для логического удаления лекарства
func (h *Handler) DeleteMedication(ctx *gin.Context) {
	// считываем значение из формы, которую мы добавим в наш шаблон
	medicationIDStr := ctx.PostForm("medication_id")
	id, err := strconv.ParseUint(medicationIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Вызов функции удаления лекарства
	err = h.Repository.DeleteMedication(uint(id))
	if err != nil {
		logrus.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// после вызова сразу произойдет обновление страницы
	ctx.Redirect(http.StatusFound, "/catalog")
}

