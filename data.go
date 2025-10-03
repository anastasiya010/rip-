package main

import (
	"os"
	"sort"
	"strings"
)

// Коллекции данных (без БД)
var medicineList []Medicine
var prescriptions map[int]Prescription

// Базовый URL для Minio (можно переопределить переменной окружения)
func minioBaseURL() string {
	if v := os.Getenv("MINIO_PUBLIC_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	// По умолчанию считаем, что статика изображений отдается через /static/img/
	return "/static/img"
}

// Хелпер для построения URL картинки из ключа
func imageURL(key string) string {
	return minioBaseURL() + "/" + key
}

// Инициализация тестовых данных
func seedData() {
	medicineList = []Medicine{
		{ID: 1, Name: "Нурофен детский", ImageKey: "nurofen.png", Category: "Жаропонижающее", Manufacturer: "Reckitt", ShortInfo: "Ибупрофен 100 мг/5 мл", Description: "Нурофен применяется при головной боли, мигрени, зубной боли, повышенной температуре, невралгии, боли в ушах, мышечной и ревматической боли."},
		{ID: 2, Name: "Пенталгин", ImageKey: "pentalgin.png", Category: "Анальгетик", Manufacturer: "OTCpharm", ShortInfo: "Комбинированный анальгетик", Description: "Пенталгин оказывает анальгезирующее и спазмолитическое действие."},
		{ID: 3, Name: "Гинкоум", ImageKey: "ginkoum.png", Category: "Ноотроп", Manufacturer: "Evalar", ShortInfo: "Экстракт гинкго билоба", Description: "Улучшает мозговое кровообращение, показан при снижении памяти и внимания."},
		{ID: 4, Name: "Цитовир-3", ImageKey: "cytovir.png", Category: "Противовирусное", Manufacturer: "Петровакс", ShortInfo: "Иммуномодулятор", Description: "Препарат с иммуномодулирующим действием для профилактики ОРВИ."},
		{ID: 5, Name: "Ринза", ImageKey: "rinza.png", Category: "От простуды", Manufacturer: "Unichem", ShortInfo: "При симптомах простуды", Description: "Комбинированное средство для снижения температуры и облегчения симптомов простуды."},
		{ID: 6, Name: "Флуимуцил", ImageKey: "fluim.png", Category: "Муколитическое", Manufacturer: "ЗАМБОН", ShortInfo: "Муколитическое средство, разжижает мокроту, увеличивает ее объем, облегчает отделение мокроты.", Description: "Флуимуцил — это муколитический (секретолитический) препарат с прямым действием. Его основная задача — разжижать мокроту и облегчать ее выведение из дыхательных путей. Он также обладает антиоксидантными свойствами."},
		{ID: 7, Name: "Бронхо-мунал", ImageKey: "bronhomun.png", Category: "Противовирусное", Manufacturer: "Sandoz", ShortInfo: "Бронхо-Мунал укрепляет иммунитет при простуде, против вирусов и бактерий, 7 мг, 10 капсул", Description: "Бронхо-Мунал — это иммуномодулирующий препарат бактериального происхождения. Он не является антибиотиком или противовирусным средством в прямом смысле. Его задача — обучить и активировать иммунную систему для борьбы с инфекциями дыхательных путей."},
	}

	// Сортируем по имени для стабильного вывода
	sort.Slice(medicineList, func(i, j int) bool { return medicineList[i].Name < medicineList[j].Name })

	// Изначально рецепт пустой
	prescriptions = map[int]Prescription{
		1001: {
			ID:         1001,
			Title:      "Рецепт на расчет дозы",
			Comment:    "",
			ResultNote: "",
			Items:      []PrescriptionItem{},
		},
	}
}

// Подсчет количества лекарств в рецепте
func prescriptionCount(p Prescription) int {
	total := 0
	for _, it := range p.Items {
		total += it.Quantity
	}
	return total
}

// Поиск лекарства по ID
func findMedicineByID(id int) (Medicine, bool) {
	for _, m := range medicineList {
		if m.ID == id {
			return m, true
		}
	}
	return Medicine{}, false
}

// Получить ID текущего рецепта (первый в словаре)
func currentPrescriptionID() int {
	for id := range prescriptions {
		return id
	}
	// если по каким-то причинам пусто — создадим дефолтный
	prescriptions = map[int]Prescription{
		1001: {ID: 1001, Title: "Рецепт на расчет дозы", Items: []PrescriptionItem{}},
	}
	return 1001
}

// Добавить лекарство в рецепт (увеличит количество, если уже есть)
func addMedicineToPrescription(medicineID int) {
	pid := currentPrescriptionID()
	p := prescriptions[pid]
	med, ok := findMedicineByID(medicineID)
	if !ok {
		return
	}
	found := false
	for i := range p.Items {
		if p.Items[i].MedicineID == medicineID {
			p.Items[i].Quantity += 1
			found = true
			break
		}
	}
	if !found {
		p.Items = append(p.Items, PrescriptionItem{
			MedicineID:   med.ID,
			MedicineName: med.Name,
			ImageKey:     med.ImageKey,
			Quantity:     1,
			Note:         "",
		})
	}
	prescriptions[pid] = p
}

// Удалить одно лекарство (уменьшить количество, удалить если 0)
func removeMedicineFromPrescription(medicineID int) {
	pid := currentPrescriptionID()
	p := prescriptions[pid]
	for i := range p.Items {
		if p.Items[i].MedicineID == medicineID {
			// Полностью удалить позицию из рецепта
			p.Items = append(p.Items[:i], p.Items[i+1:]...)
			break
		}
	}
	prescriptions[pid] = p
}

// Полная очистка рецепта
func clearPrescription() {
	pid := currentPrescriptionID()
	p := prescriptions[pid]
	p.Items = []PrescriptionItem{}
	prescriptions[pid] = p
}
