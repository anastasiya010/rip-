package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Рендер шаблона с общими данными о лекарствах
func renderPharmacyTemplate(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	base := template.Must(template.ParseFiles(
		"templates/layout.html",
		fmt.Sprintf("templates/%s.html", name),
	))
	if data == nil {
		data = map[string]any{}
	}
	data["MinioBase"] = pharmacyImageBaseURL()
	data["ImageURL"] = medicationImageURL
	if err := base.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Контроллер каталога карточек лекарств с фильтрацией по категории и производителю
func handleMedicationCatalog(w http.ResponseWriter, r *http.Request) {
	// Параметры фильтрации
	categoryFilter := strings.TrimSpace(r.URL.Query().Get("cat"))
	manufacturerFilter := strings.TrimSpace(r.URL.Query().Get("man"))

	var filteredMedicationCards []MedicationCard
	for _, medication := range medicationCatalog {
		if categoryFilter != "" && !strings.EqualFold(medication.Category, categoryFilter) {
			continue
		}
		if manufacturerFilter != "" && !strings.EqualFold(medication.Manufacturer, manufacturerFilter) {
			continue
		}
		filteredMedicationCards = append(filteredMedicationCards, medication)
	}

	// Соберём уникальные списки для селектов
	categoryDictionary := map[string]struct{}{}
	manufacturerDictionary := map[string]struct{}{}
	for _, medication := range medicationCatalog {
		categoryDictionary[medication.Category] = struct{}{}
		manufacturerDictionary[medication.Manufacturer] = struct{}{}
	}
	var categoryOptions, manufacturerOptions []string
	for category := range categoryDictionary {
		categoryOptions = append(categoryOptions, category)
	}
	for manufacturer := range manufacturerDictionary {
		manufacturerOptions = append(manufacturerOptions, manufacturer)
	}
	sort.Strings(categoryOptions)
	sort.Strings(manufacturerOptions)

	// Текущая заявка (первая в словаре для примера)
	var activePrescription PrescriptionRequest
	for _, prescription := range prescriptionRegistry {
		activePrescription = prescription
		break
	}

	renderPharmacyTemplate(w, r, "catalog", map[string]any{
		"MedicationCards":             filteredMedicationCards,
		"Prescription":                activePrescription,
		"PrescriptionMedicationCount": prescriptionMedicationCount(activePrescription),
		"SelectedCategory":            categoryFilter,
		"SelectedManufacturer":        manufacturerFilter,
		"Categories":                  categoryOptions,
		"Manufacturers":               manufacturerOptions,
	})
}

// Контроллер детали карточки лекарства: /service?id=1
func handleMedicationCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	var foundMedication *MedicationCard
	for i := range medicationCatalog {
		if medicationCatalog[i].ID == id {
			foundMedication = &medicationCatalog[i]
			break
		}
	}
	if foundMedication == nil {
		http.NotFound(w, r)
		return
	}
	renderPharmacyTemplate(w, r, "service", map[string]any{"MedicationCard": foundMedication})
}

// Контроллер заявки на назначение лекарств: /order?id=1001
func handlePrescriptionRequest(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	prescription, ok := prescriptionRegistry[id]
	if !ok {
		http.NotFound(w, r)
		return
	}
	renderPharmacyTemplate(w, r, "order", map[string]any{
		"Prescription":                prescription,
		"PrescriptionMedicationCount": prescriptionMedicationCount(prescription),
	})
}
