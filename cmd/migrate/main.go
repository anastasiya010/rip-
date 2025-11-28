package main

import (
	"fmt"
	"metoda/internal/app/ds"
	"metoda/internal/app/dsn"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Пробуем загрузить .env из разных возможных мест
	envPaths := []string{
		".env",          // Текущая директория
		"../.env",       // На уровень выше (если запускаем из cmd/migrate)
		"../../.env",    // На два уровня выше
		"../../../.env", // На три уровня выше
	}

	var err error
	loaded := false
	for _, path := range envPaths {
		err = godotenv.Load(path)
		if err == nil {
			fmt.Printf("Loaded .env from: %s\n", path)
			loaded = true
			break
		}
	}

	if !loaded {
		fmt.Printf("Warning: .env file not found, using environment variables\n")
	}

	// Выводим отладочную информацию
	dsnString := dsn.FromEnv()

	// Если DSN пустой, используем значения по умолчанию
	if dsnString == "" {
		fmt.Printf("DSN is empty, trying default values...\n")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5432")
		os.Setenv("DB_USER", "myuser")
		os.Setenv("DB_PASS", "mypassword")
		os.Setenv("DB_NAME", "medication_db")
		dsnString = dsn.FromEnv()
	}

	fmt.Printf("DSN connection string: host=%s port=%s user=%s password=*** dbname=%s\n",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_NAME"))

	if dsnString == "" {
		panic("DSN string is empty! Check .env file and DB_* variables")
	}

	db, err := gorm.Open(postgres.Open(dsnString), &gorm.Config{})
	if err != nil {
		fmt.Printf("Connection error: %v\n", err)
		panic("failed to connect database")
	}

	// Migrate the schema
	fmt.Println("Starting database migration...")
	// Важно: порядок миграции должен учитывать зависимости внешних ключей
	// User должен быть первым, так как Prescription ссылается на него
	err = db.AutoMigrate(
		&ds.User{},                   // Независимая таблица, должна быть первой
		&ds.Medication{},             // Независимая таблица
		&ds.Prescription{},           // Зависит от User
		&ds.PrescriptionMedication{}, // Зависит от Prescription и Medication
	)
	if err != nil {
		fmt.Printf("Migration error: %v\n", err)
		panic("cant migrate db")
	}

	// Создаем частичный уникальный индекс для ограничения одной заявки в статусе черновик на пользователя
	// Проверяем, существует ли уже индекс
	var indexExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes 
			WHERE indexname = 'idx_prescriptions_creator_draft_unique'
		)
	`).Scan(&indexExists).Error
	if err != nil {
		fmt.Printf("Error checking index existence: %v\n", err)
	}

	if !indexExists {
		fmt.Println("Creating partial unique index for draft prescriptions...")
		err = db.Exec(`
			CREATE UNIQUE INDEX idx_prescriptions_creator_draft_unique 
			ON prescriptions (creator_id) 
			WHERE status = 'черновик'
		`).Error
		if err != nil {
			fmt.Printf("Warning: Could not create partial unique index: %v\n", err)
			fmt.Println("This index ensures only one draft prescription per user.")
		} else {
			fmt.Println("Partial unique index created successfully!")
		}
	} else {
		fmt.Println("Partial unique index already exists.")
	}

	fmt.Println("Migration completed successfully!")
}
