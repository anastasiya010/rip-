package main

// Модель карточки лекарства
type MedicationCard struct {
	ID           int
	Name         string
	ImageKey     string // ключ в Minio
	Category     string
	Manufacturer string
	ShortInfo    string
	Description  string
}

// Модель заявки на назначение лекарств
type PrescriptionRequest struct {
	ID                  int
	RequestTitle        string
	RequestComment      string
	MedicationGuideline string // поле результата вычислений (пока фиксированное)
	Medications         []PrescribedMedication
}

// Элемент заявки c конкретным лекарством
type PrescribedMedication struct {
	MedicationCardID  int
	MedicationName    string
	ImageKey          string
	Quantity          int
	DosageInstruction string // поле м-м (комментарий/результат)
}
