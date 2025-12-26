package ds

// Medication - модель лекарственного препарата (услуга)
type Medication struct {
	ID           uint    `gorm:"primaryKey"`
	Name         string  `gorm:"type:varchar(100);not null"`          // Наименование
	Description  string  `gorm:"type:text"`                           // Описание
	IsDeleted    bool    `gorm:"type:boolean;not null;default:false"` // Статус удален/действует
	ImageURL     *string `gorm:"type:varchar(255)"`                   // URL к изображению (Nullable)
	Category     string  `gorm:"type:varchar(50)"`                    // Категория (поле по предметной области)
	Manufacturer string  `gorm:"type:varchar(100)"`                   // Производитель (поле по предметной области)
	ShortInfo    string  `gorm:"type:varchar(200)"`                   // Краткая информация (поле по предметной области)
	AdultDose    float64 `gorm:"type:decimal(10,2);default:0"`        // Доза для взрослого в мг (используется только для расчета детской дозы)
}
