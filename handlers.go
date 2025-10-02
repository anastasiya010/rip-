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

// Контроллер каталога с фильтрацией по категории и производителю
func handleCatalog(w http.ResponseWriter, r *http.Request) {
	// Параметры фильтрации
	cat := strings.TrimSpace(r.URL.Query().Get("cat"))
	man := strings.TrimSpace(r.URL.Query().Get("man"))

	var items []Service
	for _, s := range serviceList {
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
	for _, s := range serviceList {
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

	// Текущая заявка (первая в словаре для примера)
	var cur Order
	for _, o := range orders {
		cur = o
		break
	}

	render(w, r, "catalog", map[string]any{
		"Services":            items,
		"Order":               cur,
		"OrderCount":          orderCount(cur),
		"SelectedCategory":    cat,
		"SelectedManufacturer": man,
		"Categories":          catList,
		"Manufacturers":       manList,
	})
}

// Контроллер детали услуги: /service?id=1
func handleService(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	var found *Service
	for i := range serviceList {
		if serviceList[i].ID == id {
			found = &serviceList[i]
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}
	render(w, r, "service", map[string]any{"Service": found})
}

// Контроллер заявки: /order?id=1001
func handleOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	ord, ok := orders[id]
	if !ok {
		http.NotFound(w, r)
		return
	}
	render(w, r, "order", map[string]any{
		"Order":      ord,
		"OrderCount": orderCount(ord),
	})
}
