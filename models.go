package main

// Модель лекарства
type Medicine struct {
	ID           int
	Name         string
	ImageKey     string // ключ в Minio
	Category     string
	Manufacturer string
	ShortInfo    string
	Description  string
}

// Модель рецепта
type Prescription struct {
	ID         int
	Title      string
	Comment    string
	ResultNote string // поле результата вычислений (пока фиксированное)
	Items      []PrescriptionItem
}

// Элемент рецепта (плоская структура, без вложенных массивов в лекарстве)
type PrescriptionItem struct {
	MedicineID   int
	MedicineName string
	ImageKey     string
	Quantity     int
	Note         string // поле м-м (комментарий/результат)
}
