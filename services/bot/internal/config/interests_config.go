package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Константы для конфигурации интересов
const (
	// defaultPrimaryPercentage - процент основных интересов от общего количества (30%)
	defaultPrimaryPercentage = 0.3

	// defaultDirectoryPermissions - права доступа для создания директорий (0755)
	defaultDirectoryPermissions = 0755

	// defaultFilePermissions - права доступа для файлов конфигурации (0600)
	defaultFilePermissions = 0600
)

// InterestsConfig представляет конфигурацию системы интересов
type InterestsConfig struct {
	Matching       MatchingConfig            `json:"matching"`
	InterestLimits InterestLimitsConfig      `json:"interest_limits"`
	Categories     map[string]CategoryConfig `json:"categories"`
}

// MatchingConfig конфигурация для алгоритма сопоставления
type MatchingConfig struct {
	PrimaryInterestScore    int `json:"primary_interest_score"`
	AdditionalInterestScore int `json:"additional_interest_score"`
	MinCompatibilityScore   int `json:"min_compatibility_score"`
	MaxMatchesPerUser       int `json:"max_matches_per_user"`
}

// InterestLimitsConfig конфигурация лимитов интересов
type InterestLimitsConfig struct {
	MinPrimaryInterests int     `json:"min_primary_interests"`
	MaxPrimaryInterests int     `json:"max_primary_interests"`
	PrimaryPercentage   float64 `json:"primary_percentage"`
}

// CategoryConfig конфигурация категории
type CategoryConfig struct {
	DisplayOrder          int `json:"display_order"`
	MaxPrimaryPerCategory int `json:"max_primary_per_category"`
}

// LoadInterestsConfig загружает конфигурацию интересов из файла
func LoadInterestsConfig() (*InterestsConfig, error) {
	// Ищем файл конфигурации в разных местах
	configPaths := []string{
		"config/interests.json",          // из корня проекта
		"../config/interests.json",       // из services/bot/
		"../../config/interests.json",    // из services/bot/internal/
		"../../../config/interests.json", // из services/bot/internal/config/
	}

	var configPath string

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		// Если файл не найден, создаем с дефолтными значениями
		config := &InterestsConfig{
			Matching: MatchingConfig{
				PrimaryInterestScore:    3,
				AdditionalInterestScore: 1,
				MinCompatibilityScore:   5,
				MaxMatchesPerUser:       10,
			},
			InterestLimits: InterestLimitsConfig{
				MinPrimaryInterests: 1,
				MaxPrimaryInterests: 5,
				PrimaryPercentage:   defaultPrimaryPercentage, // 30% от общего количества интересов
			},
			Categories: map[string]CategoryConfig{
				"entertainment": {DisplayOrder: 1, MaxPrimaryPerCategory: 2},
				"education":     {DisplayOrder: 2, MaxPrimaryPerCategory: 2},
				"active":        {DisplayOrder: 3, MaxPrimaryPerCategory: 2},
				"creative":      {DisplayOrder: 4, MaxPrimaryPerCategory: 2},
				"social":        {DisplayOrder: 5, MaxPrimaryPerCategory: 2},
			},
		}

		return config, nil
	}

	// Читаем файл
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Парсим JSON
	var config InterestsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetInterestsConfig возвращает загруженную конфигурацию
func GetInterestsConfig() *InterestsConfig {
	config, _ := LoadInterestsConfig()
	return config
}

// SaveInterestsConfig сохраняет конфигурацию в файл
func SaveInterestsConfig(config *InterestsConfig) error {
	// Определяем путь для сохранения
	configPath := "config/interests.json"

	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(configPath), defaultDirectoryPermissions); err != nil {
		return err
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Записываем в файл
	return os.WriteFile(configPath, data, defaultFilePermissions) // Более безопасные права доступа
}
