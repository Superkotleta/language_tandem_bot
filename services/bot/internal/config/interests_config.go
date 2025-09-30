package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Константы для конфигурации интересов.
const (
	// defaultPrimaryPercentage - процент основных интересов от общего количества (30%).
	defaultPrimaryPercentage = 0.3

	// defaultDirectoryPermissions - права доступа для создания директорий (0755).
	defaultDirectoryPermissions = 0755

	// defaultFilePermissions - права доступа для файлов конфигурации (0600).
	defaultFilePermissions = 0600

	// defaultMaxMatchesPerUser - максимальное количество совпадений на пользователя.
	defaultMaxMatchesPerUser = 10

	// Константы для алгоритма сопоставления.
	defaultPrimaryInterestScore    = 3
	defaultAdditionalInterestScore = 1
	defaultMinCompatibilityScore   = 5

	// Константы для лимитов интересов.
	defaultMinPrimaryInterests = 1
	defaultMaxPrimaryInterests = 5

	// Константы для категорий.
	defaultMaxPrimaryPerCategory = 2

	// Константы для порядка отображения категорий.
	entertainmentDisplayOrder = 1
	educationDisplayOrder     = 2
	activeDisplayOrder        = 3
	creativeDisplayOrder      = 4
	socialDisplayOrder        = 5
)

// InterestsConfig представляет конфигурацию системы интересов.
type InterestsConfig struct {
	Matching       MatchingConfig            `json:"matching"`
	InterestLimits InterestLimitsConfig      `json:"interestLimits"`
	Categories     map[string]CategoryConfig `json:"categories"`
}

// MatchingConfig конфигурация для алгоритма сопоставления.
type MatchingConfig struct {
	PrimaryInterestScore    int `json:"primaryInterestScore"`
	AdditionalInterestScore int `json:"additionalInterestScore"`
	MinCompatibilityScore   int `json:"minCompatibilityScore"`
	MaxMatchesPerUser       int `json:"maxMatchesPerUser"`
}

// InterestLimitsConfig конфигурация лимитов интересов.
type InterestLimitsConfig struct {
	MinPrimaryInterests int     `json:"minPrimaryInterests"`
	MaxPrimaryInterests int     `json:"maxPrimaryInterests"`
	PrimaryPercentage   float64 `json:"primaryPercentage"`
}

// CategoryConfig конфигурация категории.
type CategoryConfig struct {
	DisplayOrder          int `json:"displayOrder"`
	MaxPrimaryPerCategory int `json:"maxPrimaryPerCategory"`
}

// LoadInterestsConfig загружает конфигурацию интересов из файла.
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
				PrimaryInterestScore:    defaultPrimaryInterestScore,
				AdditionalInterestScore: defaultAdditionalInterestScore,
				MinCompatibilityScore:   defaultMinCompatibilityScore,
				MaxMatchesPerUser:       defaultMaxMatchesPerUser,
			},
			InterestLimits: InterestLimitsConfig{
				MinPrimaryInterests: defaultMinPrimaryInterests,
				MaxPrimaryInterests: defaultMaxPrimaryInterests,
				PrimaryPercentage:   defaultPrimaryPercentage, // 30% от общего количества интересов
			},
			Categories: map[string]CategoryConfig{
				"entertainment": {DisplayOrder: entertainmentDisplayOrder, MaxPrimaryPerCategory: defaultMaxPrimaryPerCategory},
				"education":     {DisplayOrder: educationDisplayOrder, MaxPrimaryPerCategory: defaultMaxPrimaryPerCategory},
				"active":        {DisplayOrder: activeDisplayOrder, MaxPrimaryPerCategory: defaultMaxPrimaryPerCategory},
				"creative":      {DisplayOrder: creativeDisplayOrder, MaxPrimaryPerCategory: defaultMaxPrimaryPerCategory},
				"social":        {DisplayOrder: socialDisplayOrder, MaxPrimaryPerCategory: defaultMaxPrimaryPerCategory},
			},
		}

		return config, nil
	}

	// Очищаем путь для безопасности
	cleanPath := filepath.Clean(configPath)

	// Проверяем, что путь не содержит опасные символы
	if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "~") {
		return nil, fmt.Errorf("небезопасный путь к файлу: %s", configPath)
	}

	// Читаем файл
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Парсим JSON
	var config InterestsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// GetInterestsConfig возвращает загруженную конфигурацию.
func GetInterestsConfig() *InterestsConfig {
	config, _ := LoadInterestsConfig()

	return config
}

// SaveInterestsConfig сохраняет конфигурацию в файл.
func SaveInterestsConfig(config *InterestsConfig) error {
	// Определяем путь для сохранения
	configPath := "config/interests.json"

	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(configPath), defaultDirectoryPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(configPath, data, defaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
