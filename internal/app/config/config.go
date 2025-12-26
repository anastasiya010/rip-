package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
}

func NewConfig() (*Config, error) {
	var err error
	configName := "config"

	// Пробуем загрузить .env из разных возможных мест
	envPaths := []string{
		".env",          // Текущая директория
		"../.env",       // На уровень выше
		"../../.env",    // На два уровня выше
		"../../../.env", // На три уровня выше
	}

	loaded := false
	for _, path := range envPaths {
		err = godotenv.Load(path)
		if err == nil {
			log.Infof("Loaded .env from: %s", path)
			loaded = true
			break
		}
	}

	// Если не нашли .env, пробуем найти в корне проекта
	if !loaded {
		wd, _ := os.Getwd()
		// Если мы в cmd/awesomeProject, поднимаемся на уровень выше
		if strings.Contains(wd, "cmd") || strings.Contains(wd, "awesomeProject") {
			rootEnv := filepath.Join(wd, "..", "..", ".env")
			if _, err := os.Stat(rootEnv); err == nil {
				godotenv.Load(rootEnv)
				log.Infof("Loaded .env from: %s", rootEnv)
				loaded = true
			}
		}
		// Пробуем найти .env в текущей директории или выше
		if !loaded {
			for i := 0; i < 5; i++ {
				envPath := filepath.Join(wd, strings.Repeat("..", i), ".env")
				if _, err := os.Stat(envPath); err == nil {
					godotenv.Load(envPath)
					log.Infof("Loaded .env from: %s", envPath)
					loaded = true
					break
				}
			}
		}
	}

	if !loaded {
		log.Warn("Warning: .env file not found, using environment variables or defaults")
	}

	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}
	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()
	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}
	log.Info("config parsed")
	return cfg, nil
}
