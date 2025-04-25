package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	DBSouceProd          string        `mapstructure:"DB_SOURCE_PROD"`
	ENV                  string        `mapstructure:"ENV"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccesTokenDuration   time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

func LoadConfig(configName ...string) (config Config, err error) {
	fileName := "app"
	if len(configName) > 0 && configName[0] != "" {
		fileName = configName[0]
	}

	// Возможные пути к файлу конфигурации
	possiblePaths := []string{
		".",            // Текущая директория
		"./cmd",        // Относительно корня проекта или Docker
		"..",           // Родительская директория
		"../..",        // Родительская директория уровнем выше
		"../../cmd",    // cmd относительно корня проекта
		"../../../cmd", // Для очень глубоких вложенных директорий
		"../cmd",       // Для поддиректорий
	}

	// Добавляем пути в viper
	for _, path := range possiblePaths {
		viper.AddConfigPath(path)
	}

	viper.SetConfigName(fileName) // Имя файла без расширения
	viper.SetConfigType("env")    // Формат файла — .env

	// Поддержка переменных окружения (опционально)
	viper.AutomaticEnv()

	// Чтение конфигурации
	err = viper.ReadInConfig()
	if err != nil {
		return config, fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}

	// Декодирование в структуру
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, fmt.Errorf("не удалось декодировать конфигурацию: %w", err)
	}

	return config, nil
}
