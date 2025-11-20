package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	errorsPkg "language-exchange-bot/internal/errors"
	"language-exchange-bot/internal/localization"
)

// Interest configuration constants are now centralized in localization/constants.go

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
				PrimaryInterestScore:    localization.DefaultPrimaryInterestScore,
				AdditionalInterestScore: localization.DefaultAdditionalInterestScore,
				MinCompatibilityScore:   localization.DefaultMinCompatibilityScore,
				MaxMatchesPerUser:       localization.DefaultMaxMatchesPerUser,
			},
			InterestLimits: InterestLimitsConfig{
				MinPrimaryInterests: localization.DefaultMinPrimaryInterests,
				MaxPrimaryInterests: localization.DefaultMaxPrimaryInterests,
				PrimaryPercentage:   localization.DefaultPrimaryPercentage, // 30% от общего количества интересов
			},
			Categories: map[string]CategoryConfig{
				"entertainment": {DisplayOrder: localization.EntertainmentDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
				"education":     {DisplayOrder: localization.EducationDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
				"active":        {DisplayOrder: localization.ActiveDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
				"creative":      {DisplayOrder: localization.CreativeDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
				"social":        {DisplayOrder: localization.SocialDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
			},
		}

		return config, nil
	}

	// Очищаем путь для безопасности
	cleanPath := filepath.Clean(configPath)

	// Проверяем, что путь не содержит опасные символы
	if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "~") {
		return nil, errorsPkg.ErrUnsafeFilePath
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
	if config == nil {
		return errors.New("config cannot be nil")
	}

	// Определяем путь для сохранения
	configPath := "config/interests.json"

	// Создаем директорию если не существует
	if err := os.MkdirAll(filepath.Dir(configPath), localization.DefaultDirectoryPermissions); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Сериализуем в JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Записываем в файл
	if err := os.WriteFile(configPath, data, localization.DefaultFilePermissions); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
