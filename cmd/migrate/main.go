package main

import (
	"fmt"
	"metoda/internal/app/ds"
	"metoda/internal/app/dsn"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
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

	// Если не нашли .env, пробуем загрузить из корня проекта
	if !loaded {
		wd, _ := os.Getwd()
		// Если мы в cmd/migrate, поднимаемся на уровень выше
		if strings.Contains(wd, "cmd") || strings.Contains(wd, "migrate") {
			rootEnv := filepath.Join(wd, "..", "..", ".env")
			if _, err := os.Stat(rootEnv); err == nil {
				godotenv.Load(rootEnv)
				fmt.Printf("Loaded .env from: %s\n", rootEnv)
				loaded = true
			}
		}
		// Пробуем найти .env в текущей директории или выше
		if !loaded {
			for i := 0; i < 5; i++ {
				envPath := filepath.Join(wd, strings.Repeat("..", i), ".env")
				if _, err := os.Stat(envPath); err == nil {
					godotenv.Load(envPath)
					fmt.Printf("Loaded .env from: %s\n", envPath)
					loaded = true
					break
				}
			}
		}
	}

	if !loaded {
		fmt.Printf("Warning: .env file not found, using environment variables or defaults\n")
	}

	// Выводим отладочную информацию
	dsnString := dsn.FromEnv()

	// Если DSN пустой, используем значения по умолчанию
	if dsnString == "" {
		fmt.Printf("DSN is empty, trying default values...\n")
		os.Setenv("DB_HOST", "localhost")
		os.Setenv("DB_PORT", "5434") // Порт из docker-compose.yml
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

	// Загружаем данные из seed_data.sql
	fmt.Println("Loading seed data from sql/seed_data.sql...")
	err = loadSeedData(db)
	if err != nil {
		fmt.Printf("Warning: Could not load seed data: %v\n", err)
		fmt.Println("You can manually load data from sql/seed_data.sql")
	} else {
		fmt.Println("Seed data loaded successfully!")
	}

	// Обновляем пароли пользователей, если они еще не захешированы
	fmt.Println("Updating user passwords with bcrypt hashes...")
	err = updateUserPasswords(db)
	if err != nil {
		fmt.Printf("Warning: Could not update passwords: %v\n", err)
	} else {
		fmt.Println("User passwords updated successfully!")
	}

	fmt.Println("Migration completed successfully!")
}

// loadSeedData загружает данные из файла seed_data.sql
func loadSeedData(db *gorm.DB) error {
	// Пробуем найти файл seed_data.sql в разных местах
	seedPaths := []string{
		"sql/seed_data.sql",
		"../sql/seed_data.sql",
		"../../sql/seed_data.sql",
		"../../../sql/seed_data.sql",
	}

	var seedData []byte
	var err error
	var seedPath string

	for _, path := range seedPaths {
		seedData, err = os.ReadFile(path)
		if err == nil {
			seedPath = path
			fmt.Printf("Found seed file at: %s\n", path)
			break
		}
	}

	if err != nil {
		// Пробуем найти файл относительно текущей директории
		wd, _ := os.Getwd()
		possiblePath := filepath.Join(wd, "sql", "seed_data.sql")
		seedData, err = os.ReadFile(possiblePath)
		if err == nil {
			seedPath = possiblePath
			fmt.Printf("Found seed file at: %s\n", possiblePath)
		}
	}

	if err != nil {
		return fmt.Errorf("seed_data.sql file not found: %v", err)
	}

	// Разбиваем SQL на отдельные запросы
	// Учитываем, что точка с запятой может быть внутри строк
	sqlContent := string(seedData)

	// Удаляем комментарии (строки, начинающиеся с --)
	lines := strings.Split(sqlContent, "\n")
	var cleanedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Пропускаем пустые строки и комментарии
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}

	// Объединяем обратно и разбиваем по точкам с запятой
	cleanedSQL := strings.Join(cleanedLines, "\n")
	sqlStatements := strings.Split(cleanedSQL, ";")

	// Выполняем каждый SQL запрос
	executedCount := 0
	for i, stmt := range sqlStatements {
		// Убираем пробелы и переносы строк
		stmt = strings.TrimSpace(stmt)

		// Пропускаем пустые строки
		if stmt == "" {
			continue
		}

		// Выполняем запрос
		err = db.Exec(stmt).Error
		if err != nil {
			// Игнорируем ошибки о дублировании ключей (данные уже могут быть загружены)
			errStr := err.Error()
			if strings.Contains(errStr, "duplicate key") ||
				strings.Contains(errStr, "already exists") ||
				strings.Contains(errStr, "violates unique constraint") ||
				strings.Contains(errStr, "unique constraint") {
				fmt.Printf("Skipping statement %d (data already exists)\n", i+1)
				continue
			}
			return fmt.Errorf("error executing statement %d: %v\nStatement preview: %s", i+1, err, stmt[:min(100, len(stmt))])
		}
		executedCount++
	}

	fmt.Printf("Executed %d SQL statements from %s\n", executedCount, seedPath)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// updateUserPasswords обновляет пароли пользователей, хешируя их через bcrypt
func updateUserPasswords(db *gorm.DB) error {
	// Список пользователей с паролями
	users := map[string]string{
		"user1": "user1123",
		"user2": "user2123",
		"admin": "admin123",
	}

	for login, password := range users {
		var user ds.User
		err := db.Where("login = ?", login).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				fmt.Printf("User %s not found, skipping password update\n", login)
				continue // Пользователь не найден, пропускаем
			}
			return fmt.Errorf("error finding user %s: %w", login, err)
		}

		// Проверяем, не захеширован ли уже пароль
		if strings.HasPrefix(user.Password, "$2a$") || strings.HasPrefix(user.Password, "$2b$") || strings.HasPrefix(user.Password, "$2y$") {
			// Пароль уже захеширован, пропускаем
			fmt.Printf("Password for user %s is already hashed, skipping\n", login)
			continue
		}

		// Хешируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error hashing password for %s: %w", login, err)
		}

		// Обновляем пароль
		err = db.Model(&user).Update("password", string(hashedPassword)).Error
		if err != nil {
			return fmt.Errorf("error updating password for %s: %w", login, err)
		}

		fmt.Printf("Password updated for user: %s\n", login)
	}

	return nil
}
