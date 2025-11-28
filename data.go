package main

import (
	"os"
	"sort"
	"strings"
)

// Коллекции данных (без БД)
var medicationCatalog []MedicationCard
var prescriptionRegistry map[int]PrescriptionRequest

// Базовый URL для хранилища изображений лекарств (можно переопределить переменной окружения)
func pharmacyImageBaseURL() string {
	if v := os.Getenv("MINIO_PUBLIC_BASE"); v != "" {
		return strings.TrimRight(v, "/")
	}
	// По умолчанию считаем, что статика изображений отдается через /static/img/
	return "/static/img"
}

// Хелпер для построения URL картинки лекарства из ключа
func medicationImageURL(key string) string {
	return pharmacyImageBaseURL() + "/" + key
}

// Инициализация тестовых данных о лекарствах
func seedMedicationDataset() {
	medicationCatalog = []MedicationCard{
		{ID: 1, Name: "Нурофен детский", ImageKey: "nurofen.png", Category: "Жаропонижающее", Manufacturer: "Reckitt", ShortInfo: "Ибупрофен 100 мг/5 мл", Description: "Нурофен применяется при головной боли, мигрени, зубной боли, повышенной температуре, невралгии, боли в ушах, мышечной и ревматической боли."},
		{ID: 2, Name: "Пенталгин", ImageKey: "pentalgin.png", Category: "Анальгетик", Manufacturer: "OTCpharm", ShortInfo: "Комбинированный анальгетик", Description: "Пенталгин оказывает анальгезирующее и спазмолитическое действие."},
		{ID: 3, Name: "Гинкоум", ImageKey: "ginkoum.png", Category: "Ноотроп", Manufacturer: "Evalar", ShortInfo: "Экстракт гинкго билоба", Description: "Улучшает мозговое кровообращение, показан при снижении памяти и внимания."},
		{ID: 4, Name: "Цитовир-3", ImageKey: "cytovir.png", Category: "Противовирусное", Manufacturer: "Петровакс", ShortInfo: "Иммуномодулятор", Description: "Препарат с иммуномодулирующим действием для профилактики ОРВИ."},
		{ID: 5, Name: "Ринза", ImageKey: "rinza.png", Category: "От простуды", Manufacturer: "Unichem", ShortInfo: "При симптомах простуды", Description: "Комбинированное средство для снижения температуры и облегчения симптомов простуды."},
		{ID: 6, Name: "Флуимуцил", ImageKey: "fluim.png", Category: "Муколитическое", Manufacturer: "ЗАМБОН", ShortInfo: "Муколитическое средство, разжижает мокроту, увеличивает ее объем, облегчает отделение мокроты.", Description: "Флуимуцил — это муколитический (секретолитический) препарат с прямым действием. Его основная задача — разжижать мокроту и облегчать ее выведение из дыхательных путей. Он также обладает антиоксидантными свойствами."},
		{ID: 7, Name: "Бронхо-мунал", ImageKey: "bronhomun.png", Category: "Противовирусное", Manufacturer: "Sandoz", ShortInfo: "Бронхо-Мунал укрепляет иммунитет при простуде, против вирусов и бактерий, 7 мг, 10 капсул", Description: "Бронхо-Мунал — это иммуномодулирующий препарат бактериального происхождения. Он не является антибиотиком или противовирусным средством в прямом смысле. Его задача — обучить и активировать иммунную систему для борьбы с инфекциями дыхательных путей."},
	}

	// Сортируем по имени для стабильного вывода
	sort.Slice(medicationCatalog, func(i, j int) bool { return medicationCatalog[i].Name < medicationCatalog[j].Name })

	prescriptionRegistry = map[int]PrescriptionRequest{
		1001: {
			ID:                  1001,
			RequestTitle:        "Заявка на расчет дозы",
			RequestComment:      "Пациент: ребенок 6 лет",
			MedicationGuideline: "Итог: расчет дозы произведен врачом. Рекомендация: Нурофен 10 мг/кг",
			Medications: []PrescribedMedication{
				{MedicationCardID: 2, MedicationName: "Пенталгин", ImageKey: "pentalgin.png", Quantity: 1, DosageInstruction: "по необходимости"},
				{MedicationCardID: 3, MedicationName: "Гинкоум", ImageKey: "ginkoum.png", Quantity: 1, DosageInstruction: "курс 30 дней"},
			},
		},
	}
}

// Подсчет количества назначенных лекарств в заявке
func prescriptionMedicationCount(p PrescriptionRequest) int {
	total := 0
	for _, medication := range p.Medications {
		total += medication.Quantity
	}
	return total
}
