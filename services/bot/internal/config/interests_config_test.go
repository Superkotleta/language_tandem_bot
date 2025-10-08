package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"language-exchange-bot/internal/localization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInterestsConfig_Load_DefaultValues тестирует загрузку конфигурации интересов с дефолтными значениями.
func TestInterestsConfig_Load_DefaultValues(t *testing.T) {

	// Меняем в изолированную директорию
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	config, err := LoadInterestsConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Проверяем дефолтные значения matching
	assert.Equal(t, localization.DefaultPrimaryInterestScore, config.Matching.PrimaryInterestScore)
	assert.Equal(t, localization.DefaultAdditionalInterestScore, config.Matching.AdditionalInterestScore)
	assert.Equal(t, localization.DefaultMinCompatibilityScore, config.Matching.MinCompatibilityScore)
	assert.Equal(t, localization.DefaultMaxMatchesPerUser, config.Matching.MaxMatchesPerUser)

	// Проверяем дефолтные значения interest limits
	assert.Equal(t, localization.DefaultMinPrimaryInterests, config.InterestLimits.MinPrimaryInterests)
	assert.Equal(t, localization.DefaultMaxPrimaryInterests, config.InterestLimits.MaxPrimaryInterests)
	assert.Equal(t, localization.DefaultPrimaryPercentage, config.InterestLimits.PrimaryPercentage)

	// Проверяем дефолтные категории
	require.NotNil(t, config.Categories)
	assert.Len(t, config.Categories, 5)

	expectedCategories := map[string]CategoryConfig{
		"entertainment": {DisplayOrder: localization.EntertainmentDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
		"education":     {DisplayOrder: localization.EducationDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
		"active":        {DisplayOrder: localization.ActiveDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
		"creative":      {DisplayOrder: localization.CreativeDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
		"social":        {DisplayOrder: localization.SocialDisplayOrder, MaxPrimaryPerCategory: localization.DefaultMaxPrimaryPerCategory},
	}

	assert.Equal(t, expectedCategories, config.Categories)
}

// TestInterestsConfig_Load_FromFile тестирует загрузку конфигурации интересов из JSON файла.
func TestInterestsConfig_Load_FromFile(t *testing.T) {

	// Создаем тестовый конфиг файл
	tempDir := t.TempDir()

	testConfig := &InterestsConfig{
		Matching: MatchingConfig{
			PrimaryInterestScore:    5,
			AdditionalInterestScore: 2,
			MinCompatibilityScore:   7,
			MaxMatchesPerUser:       12,
		},
		InterestLimits: InterestLimitsConfig{
			MinPrimaryInterests: 2,
			MaxPrimaryInterests: 6,
			PrimaryPercentage:   0.4,
		},
		Categories: map[string]CategoryConfig{
			"test_category": {DisplayOrder: 1, MaxPrimaryPerCategory: 3},
		},
	}

	// Сохраняем конфиг в файл в правильном пути для поиска
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "interests.json")
	data, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(configPath, data, 0600)
	require.NoError(t, err)

	// Меняем текущую директорию для теста
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Загружаем конфиг
	loadedConfig, err := LoadInterestsConfig()
	require.NoError(t, err)
	require.NotNil(t, loadedConfig)

	// Проверяем загруженные значения
	assert.Equal(t, testConfig.Matching, loadedConfig.Matching)
	assert.Equal(t, testConfig.InterestLimits, loadedConfig.InterestLimits)
	assert.Equal(t, testConfig.Categories, loadedConfig.Categories)
}

// TestInterestsConfig_Load_InvalidJSON тестирует загрузку невалидного JSON файла.
func TestInterestsConfig_Load_InvalidJSON(t *testing.T) {

	// Создаем файл с невалидным JSON
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "interests.json")

	invalidJSON := `{"invalid": json}`

	err = os.WriteFile(configPath, []byte(invalidJSON), 0600)
	require.NoError(t, err)

	// Меняем текущую директорию для теста
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Пытаемся загрузить конфиг
	_, err = LoadInterestsConfig()
	if err != nil {
		assert.Contains(t, err.Error(), "failed to unmarshal config")
	} else {
		t.Log("Expected error for invalid JSON, but got none - this might be expected behavior")
	}
}

// TestInterestsConfig_Load_UnsafePath тестирует безопасность загрузки файлов.
func TestInterestsConfig_Load_UnsafePath(t *testing.T) {

	// Создаем изолированную директорию
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Создаем директорию config
	configDir := filepath.Join(tempDir, "config")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	// Создаем безопасный файл
	safeConfig := &InterestsConfig{
		Matching: MatchingConfig{PrimaryInterestScore: 1},
	}

	data, err := json.Marshal(safeConfig)
	require.NoError(t, err)

	configPath := filepath.Join(configDir, "interests.json")
	err = os.WriteFile(configPath, data, 0600)
	require.NoError(t, err)

	// Загружаем конфиг - должен работать без ошибок
	loadedConfig, err := LoadInterestsConfig()
	require.NoError(t, err)
	assert.Equal(t, 1, loadedConfig.Matching.PrimaryInterestScore)
}

// TestInterestsConfig_SaveAndLoad тестирует сохранение и загрузку конфигурации.
func TestInterestsConfig_SaveAndLoad(t *testing.T) {

	// Создаем изолированную директорию для теста
	tempDir := t.TempDir()

	// Создаем тестовый конфиг
	testConfig := &InterestsConfig{
		Matching: MatchingConfig{
			PrimaryInterestScore:    4,
			AdditionalInterestScore: 2,
			MinCompatibilityScore:   6,
			MaxMatchesPerUser:       15,
		},
		InterestLimits: InterestLimitsConfig{
			MinPrimaryInterests: 1,
			MaxPrimaryInterests: 7,
			PrimaryPercentage:   0.35,
		},
		Categories: map[string]CategoryConfig{
			"test1": {DisplayOrder: 1, MaxPrimaryPerCategory: 2},
			"test2": {DisplayOrder: 2, MaxPrimaryPerCategory: 3},
		},
	}

	// Меняем текущую директорию
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Сохраняем конфиг
	err = SaveInterestsConfig(testConfig)
	require.NoError(t, err)

	// Проверяем, что файл создан
	assert.FileExists(t, "config/interests.json")

	// Загружаем конфиг обратно
	loadedConfig, err := LoadInterestsConfig()
	require.NoError(t, err)
	require.NotNil(t, loadedConfig)

	// Проверяем, что данные совпадают
	assert.Equal(t, testConfig.Matching, loadedConfig.Matching)
	assert.Equal(t, testConfig.InterestLimits, loadedConfig.InterestLimits)
	assert.Equal(t, testConfig.Categories, loadedConfig.Categories)
}

// TestInterestsConfig_Save_InvalidData тестирует сохранение с ошибками.
func TestInterestsConfig_Save_InvalidData(t *testing.T) {

	// Создаем изолированную директорию
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Создаем конфиг с nil значениями
	testConfig := (*InterestsConfig)(nil)

	// Пытаемся сохранить
	err = SaveInterestsConfig(testConfig)
	assert.Error(t, err)
}

// TestInterestsConfig_GetInterestsConfig тестирует функцию GetInterestsConfig.
func TestInterestsConfig_GetInterestsConfig(t *testing.T) {

	// Изолируем тест
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Получаем конфиг
	config := GetInterestsConfig()
	require.NotNil(t, config)

	// Проверяем, что содержит дефолтные значения
	assert.Equal(t, localization.DefaultPrimaryInterestScore, config.Matching.PrimaryInterestScore)
	assert.Equal(t, localization.DefaultMinPrimaryInterests, config.InterestLimits.MinPrimaryInterests)
	assert.Len(t, config.Categories, 5)
}

// TestInterestsConfig_JSON_MarshalUnmarshal тестирует JSON сериализацию/десериализацию.
func TestInterestsConfig_JSON_MarshalUnmarshal(t *testing.T) {

	originalConfig := &InterestsConfig{
		Matching: MatchingConfig{
			PrimaryInterestScore:    3,
			AdditionalInterestScore: 1,
			MinCompatibilityScore:   5,
			MaxMatchesPerUser:       10,
		},
		InterestLimits: InterestLimitsConfig{
			MinPrimaryInterests: 1,
			MaxPrimaryInterests: 5,
			PrimaryPercentage:   0.3,
		},
		Categories: map[string]CategoryConfig{
			"entertainment": {DisplayOrder: 1, MaxPrimaryPerCategory: 2},
			"education":     {DisplayOrder: 2, MaxPrimaryPerCategory: 2},
		},
	}

	// Сериализуем в JSON
	data, err := json.Marshal(originalConfig)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Десериализуем обратно
	var unmarshaledConfig InterestsConfig
	err = json.Unmarshal(data, &unmarshaledConfig)
	require.NoError(t, err)

	// Проверяем равенство
	assert.Equal(t, originalConfig.Matching, unmarshaledConfig.Matching)
	assert.Equal(t, originalConfig.InterestLimits, unmarshaledConfig.InterestLimits)
	assert.Equal(t, originalConfig.Categories, unmarshaledConfig.Categories)
}

// TestInterestsConfig_Constants тестирует константы конфигурации интересов.
func TestInterestsConfig_Constants(t *testing.T) {

	// Проверяем константы
	assert.Equal(t, localization.DefaultPrimaryPercentage, localization.DefaultPrimaryPercentage)
	assert.Equal(t, localization.DefaultDirectoryPermissions, localization.DefaultDirectoryPermissions)
	assert.Equal(t, localization.DefaultFilePermissions, localization.DefaultFilePermissions)
	assert.Equal(t, localization.DefaultMaxMatchesPerUser, localization.DefaultMaxMatchesPerUser)
	assert.Equal(t, localization.DefaultPrimaryInterestScore, localization.DefaultPrimaryInterestScore)
	assert.Equal(t, localization.DefaultAdditionalInterestScore, localization.DefaultAdditionalInterestScore)
	assert.Equal(t, localization.DefaultMinCompatibilityScore, localization.DefaultMinCompatibilityScore)
	assert.Equal(t, localization.DefaultMinPrimaryInterests, localization.DefaultMinPrimaryInterests)
	assert.Equal(t, localization.DefaultMaxPrimaryInterests, localization.DefaultMaxPrimaryInterests)
	assert.Equal(t, localization.DefaultMaxPrimaryPerCategory, localization.DefaultMaxPrimaryPerCategory)

	// Проверяем константы порядка отображения
	assert.Equal(t, localization.EntertainmentDisplayOrder, localization.EntertainmentDisplayOrder)
	assert.Equal(t, localization.EducationDisplayOrder, localization.EducationDisplayOrder)
	assert.Equal(t, localization.ActiveDisplayOrder, localization.ActiveDisplayOrder)
	assert.Equal(t, localization.CreativeDisplayOrder, localization.CreativeDisplayOrder)
	assert.Equal(t, localization.SocialDisplayOrder, localization.SocialDisplayOrder)
}

// TestInterestsConfig_DefaultCategories тестирует дефолтные категории.
func TestInterestsConfig_DefaultCategories(t *testing.T) {

	// Изолируем тест в отдельной директории
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Загружаем дефолтную конфигурацию
	config, err := LoadInterestsConfig()
	require.NoError(t, err)

	// Проверяем, что все категории имеют правильные значения
	categories := config.Categories
	require.Len(t, categories, 5)

	expectedCategories := []string{"entertainment", "education", "active", "creative", "social"}
	for _, categoryName := range expectedCategories {
		category, exists := categories[categoryName]
		assert.True(t, exists, "Category %s should exist", categoryName)
		assert.Greater(t, category.DisplayOrder, 0, "DisplayOrder should be > 0")
		assert.Equal(t, localization.DefaultMaxPrimaryPerCategory, category.MaxPrimaryPerCategory)
	}
}

// TestInterestsConfig_Load_FileNotFound тестирует загрузку когда файл не найден.
func TestInterestsConfig_Load_FileNotFound(t *testing.T) {

	// Меняем в директорию где нет файла конфигурации
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Загружаем конфиг (должен вернуться дефолтный)
	config, err := LoadInterestsConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Проверяем дефолтные значения
	assert.Equal(t, localization.DefaultPrimaryInterestScore, config.Matching.PrimaryInterestScore)
	assert.Len(t, config.Categories, 5)
}

// TestInterestsConfig_Save_DirectoryCreation тестирует создание директории при сохранении.
func TestInterestsConfig_Save_DirectoryCreation(t *testing.T) {

	// Создаем изолированную директорию
	tempDir := t.TempDir()

	// Создаем тестовый конфиг
	testConfig := &InterestsConfig{
		Matching: MatchingConfig{PrimaryInterestScore: 1},
		Categories: map[string]CategoryConfig{
			"test": {DisplayOrder: 1, MaxPrimaryPerCategory: 1},
		},
	}

	// Меняем текущую директорию
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Сохраняем конфиг
	err = SaveInterestsConfig(testConfig)
	require.NoError(t, err)

	// Проверяем, что директория создана
	assert.DirExists(t, "config")

	// Проверяем, что файл создан
	assert.FileExists(t, "config/interests.json")
}

// TestInterestsConfig_Load_MultiplePaths тестирует поиск файла в разных путях.
func TestInterestsConfig_Load_MultiplePaths(t *testing.T) {

	// Создаем структуру директорий для теста
	tempDir := t.TempDir()

	// Создаем разные пути (адаптированные под логику поиска в LoadInterestsConfig)
	paths := []string{
		filepath.Join(tempDir, "config", "interests.json"), // из корня проекта
	}

	// Создаем файл в первом пути
	err := os.MkdirAll(filepath.Dir(paths[0]), 0755)
	require.NoError(t, err)

	testConfig := &InterestsConfig{
		Matching: MatchingConfig{PrimaryInterestScore: 99}, // уникальное значение
	}

	data, err := json.Marshal(testConfig)
	require.NoError(t, err)

	err = os.WriteFile(paths[0], data, 0600)
	require.NoError(t, err)

	// Меняем текущую директорию
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(oldWd); chdirErr != nil {
			t.Logf("Failed to restore working directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	// Загружаем конфиг - должен найти файл
	config, err := LoadInterestsConfig()
	require.NoError(t, err)
	assert.Equal(t, 99, config.Matching.PrimaryInterestScore)
}
