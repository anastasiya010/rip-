package ds

import (
	"database/sql"
	"time"
)

// Prescription - модель заявки на назначение лекарств (рецепт)
type Prescription struct {
	ID                uint         `gorm:"primaryKey"`
	Status            string       `gorm:"type:varchar(20);not null"` // Статус: черновик, удалён, сформирован, завершён, отклонён
	DateCreate        time.Time    `gorm:"not null"`                  // Дата создания
	CreatorID         uint         `gorm:"not null"`                  // Создатель
	DateFormation     sql.NullTime `gorm:"default:null"`              // Дата формирования (2 действия создателя)
	DateFinish        sql.NullTime `gorm:"default:null"`              // Дата завершения (2 действия модератора)
	ModeratorID       *uint        `gorm:"default:null"`              // Модератор
	PatientName       string       `gorm:"type:varchar(100)"`         // Имя ребенка (поле по предметной области)
	PatientGender     string       `gorm:"type:varchar(10)"`          // Пол ребенка (поле по предметной области)
	PatientWeight     float64      `gorm:"type:decimal(5,2)"`         // Вес ребенка (поле по предметной области)
	PatientAge        int          `gorm:"type:integer"`              // Возраст ребенка (поле по предметной области)
	CalculationResult string       `gorm:"type:text"`                 // Результат расчета дозы (рассчитывается при завершении)

	Creator   User  `gorm:"foreignKey:CreatorID;constraint:OnDelete:NO ACTION"`
	Moderator *User `gorm:"foreignKey:ModeratorID;constraint:OnDelete:NO ACTION"`
}
