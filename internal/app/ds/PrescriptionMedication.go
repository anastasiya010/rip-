package ds

// PrescriptionMedication - модель связи многие-ко-многим между заявками и лекарствами
type PrescriptionMedication struct {
	ID            uint `gorm:"primaryKey"`
	PrescriptionID uint `gorm:"not null;uniqueIndex:idx_prescription_medication"` // Составной уникальный ключ
	MedicationID   uint `gorm:"not null;uniqueIndex:idx_prescription_medication"` // Составной уникальный ключ
	Quantity       int  `gorm:"type:integer;default:1"`                           // Количество
	OrderNumber    int  `gorm:"type:integer;default:0"`                           // Порядок
	IsMain         bool `gorm:"type:boolean;default:false"`                       // Главный препарат
	DosageInstruction string `gorm:"type:text"`                                   // Инструкция по дозировке (дополнительное поле)
	
	Prescription Prescription `gorm:"foreignKey:PrescriptionID;constraint:OnDelete:NO ACTION"`
	Medication   Medication   `gorm:"foreignKey:MedicationID;constraint:OnDelete:NO ACTION"`
}

