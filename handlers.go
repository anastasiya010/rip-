package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// Рендер шаблона с общими данными
func render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	base := template.Must(template.ParseFiles(
		"templates/layout.html",
		fmt.Sprintf("templates/%s.html", name),
	))
	if data == nil {
		data = map[string]any{}
	}
	data["MinioBase"] = minioBaseURL()
	data["ImageURL"] = imageURL
	if err := base.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Контроллер каталога лекарств с фильтрацией по категории и производителю
func handleMedicineCatalog(w http.ResponseWriter, r *http.Request) {
	// Параметры фильтрации
	cat := strings.TrimSpace(r.URL.Query().Get("cat"))
	man := strings.TrimSpace(r.URL.Query().Get("man"))

	var items []Medicine
	for _, s := range medicineList {
		if cat != "" && !strings.EqualFold(s.Category, cat) {
			continue
		}
		if man != "" && !strings.EqualFold(s.Manufacturer, man) {
			continue
		}
		items = append(items, s)
	}

	// Соберём уникальные списки для селектов
	cats := map[string]struct{}{}
	mans := map[string]struct{}{}
	for _, s := range medicineList {
		cats[s.Category] = struct{}{}
		mans[s.Manufacturer] = struct{}{}
	}
	var catList, manList []string
	for k := range cats {
		catList = append(catList, k)
	}
	for k := range mans {
		manList = append(manList, k)
	}
	sort.Strings(catList)
	sort.Strings(manList)

	// Текущий рецепт (первый в словаре для примера)
	var cur Prescription
	for _, p := range prescriptions {
		cur = p
		break
	}

	render(w, r, "catalog", map[string]any{
		"Medicines":            items,
		"Prescription":         cur,
		"PrescriptionCount":    prescriptionCount(cur),
		"SelectedCategory":     cat,
		"SelectedManufacturer": man,
		"Categories":           catList,
		"Manufacturers":        manList,
	})
}

// Контроллер детали лекарства: /medicine?id=1
func handleMedicine(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	var found *Medicine
	for i := range medicineList {
		if medicineList[i].ID == id {
			found = &medicineList[i]
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	render(w, r, "service", map[string]any{"Medicine": found})
}

// Контроллер рецепта: /prescription?id=1001
func handlePrescription(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	presc, ok := prescriptions[id]
	if !ok {
		http.NotFound(w, r)
		return
	}
	render(w, r, "order", map[string]any{
		"Prescription":      presc,
		"PrescriptionCount": prescriptionCount(presc),
	})
}

func redirectBack(w http.ResponseWriter, r *http.Request, fallback string) {
	ref := r.Header.Get("Referer")
	if ref == "" {
		http.Redirect(w, r, fallback, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}

// POST /cart/add?id=MED_ID
func handleCartAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	addMedicineToPrescription(id)
	// редирект обратно на предыдущую страницу или в каталог
	redirectBack(w, r, "/medicine-catalog")
}

// POST /cart/remove?id=MED_ID
func handleCartRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	removeMedicineFromPrescription(id)
	// редирект обратно на предыдущую страницу или в рецепт
	redirectBack(w, r, "/prescription?id="+strconv.Itoa(currentPrescriptionID()))
}

// POST /cart/clear
func handleCartClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clearPrescription()
	redirectBack(w, r, "/prescription?id="+strconv.Itoa(currentPrescriptionID()))
}
