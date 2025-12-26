package dsn

import (
	"fmt"
	"os"
)

func FromEnv() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	// Если переменные не установлены, используем значения по умолчанию
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5434" // Порт из docker-compose.yml
	}
	if user == "" {
		user = "myuser"
	}
	if pass == "" {
		pass = "mypassword"
	}
	if dbname == "" {
		dbname = "medication_db"
	}

	// Используем стандартный формат DSN для PostgreSQL с дополнительными параметрами
	// для решения проблем с аутентификацией
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable connect_timeout=10", host, port, user, pass, dbname)
}
