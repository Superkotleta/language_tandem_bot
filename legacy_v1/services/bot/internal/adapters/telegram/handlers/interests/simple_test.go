package interests

import (
	"testing"
	"time"

	"language-exchange-bot/internal/models"
)

// TestEditSession создает и тестирует сессию редактирования.
func TestEditSession(t *testing.T) {
	session := &EditSession{
		UserID:             123,
		OriginalSelections: []models.InterestSelection{},
		CurrentSelections:  []models.InterestSelection{},
		Changes:            []InterestChange{},
		CurrentCategory:    "entertainment",
		SessionStart:       time.Now(),
		LastActivity:       time.Now(),
	}

	if session.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", session.UserID)
	}

	if session.CurrentCategory != "entertainment" {
		t.Errorf("Expected category 'entertainment', got '%s'", session.CurrentCategory)
	}

	// Проверяем инициализацию коллекций
	if len(session.OriginalSelections) != 0 {
		t.Errorf("Expected empty OriginalSelections, got %d items", len(session.OriginalSelections))
	}

	if len(session.CurrentSelections) != 0 {
		t.Errorf("Expected empty CurrentSelections, got %d items", len(session.CurrentSelections))
	}

	if len(session.Changes) != 0 {
		t.Errorf("Expected empty Changes, got %d items", len(session.Changes))
	}

	// Проверяем, что время инициализировано
	if session.SessionStart.IsZero() {
		t.Error("SessionStart should not be zero")
	}

	if session.LastActivity.IsZero() {
		t.Error("LastActivity should not be zero")
	}
}

// TestInterestChange тестирует структуру изменений.
func TestInterestChange(t *testing.T) {
	change := InterestChange{
		Action:       "add",
		InterestID:   456,
		InterestName: "Кино",
		Category:     "entertainment",
		Timestamp:    time.Now(),
	}

	if change.Action != "add" {
		t.Errorf("Expected action 'add', got '%s'", change.Action)
	}

	if change.InterestID != 456 {
		t.Errorf("Expected InterestID 456, got %d", change.InterestID)
	}

	if change.InterestName != "Кино" {
		t.Errorf("Expected InterestName 'Кино', got '%s'", change.InterestName)
	}

	if change.Category != "entertainment" {
		t.Errorf("Expected Category 'entertainment', got '%s'", change.Category)
	}

	if change.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
}

// TestEditStats тестирует статистику редактирования.
func TestEditStats(t *testing.T) {
	stats := EditStats{
		TotalSelected:  5,
		PrimaryCount:   2,
		CategoryCounts: map[string]int{"entertainment": 3, "sports": 2},
		ChangesCount:   10,
		LastUpdated:    time.Now(),
	}

	if stats.TotalSelected != 5 {
		t.Errorf("Expected TotalSelected 5, got %d", stats.TotalSelected)
	}

	if stats.PrimaryCount != 2 {
		t.Errorf("Expected PrimaryCount 2, got %d", stats.PrimaryCount)
	}

	if len(stats.CategoryCounts) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(stats.CategoryCounts))
	}

	if stats.ChangesCount != 10 {
		t.Errorf("Expected ChangesCount 10, got %d", stats.ChangesCount)
	}

	if stats.LastUpdated.IsZero() {
		t.Error("LastUpdated should not be zero")
	}
}

// TestValidation тестирует валидацию.
func TestValidation(t *testing.T) {
	// Тест пустых выборов (теперь разрешено)
	emptySession := &EditSession{
		CurrentSelections: []models.InterestSelection{},
	}

	editor := &IsolatedInterestEditor{}

	err := editor.validateSelections(emptySession)
	if err != nil {
		t.Error("Empty selections should now be allowed")
	}

	// Тест с валидными выборами
	validSession := &EditSession{
		CurrentSelections: []models.InterestSelection{
			{InterestID: 1, IsPrimary: false},
			{InterestID: 2, IsPrimary: true},
		},
	}

	err = editor.validateSelections(validSession)
	if err != nil {
		t.Errorf("Unexpected validation error: %v", err)
	}
}

// TestStatisticsCalculation тестирует расчет статистики.
func TestStatisticsCalculation(t *testing.T) {
	editor := &IsolatedInterestEditor{}

	session := &EditSession{
		CurrentSelections: []models.InterestSelection{
			{InterestID: 1, IsPrimary: true},
			{InterestID: 2, IsPrimary: false},
			{InterestID: 3, IsPrimary: true},
		},
		Changes: []InterestChange{
			{Action: "add", InterestID: 1},
			{Action: "remove", InterestID: 2},
		},
	}

	stats := editor.calculateEditStats(session)

	if stats.TotalSelected != 3 {
		t.Errorf("Expected TotalSelected 3, got %d", stats.TotalSelected)
	}

	if stats.PrimaryCount != 2 {
		t.Errorf("Expected PrimaryCount 2, got %d", stats.PrimaryCount)
	}

	if stats.ChangesCount != 2 {
		t.Errorf("Expected ChangesCount 2, got %d", stats.ChangesCount)
	}
}

// TestBasicPerformance тестирует базовую производительность.
func TestBasicPerformance(t *testing.T) {
	start := time.Now()

	// Тест создания большого количества сессий
	for i := range 1000 {
		session := &EditSession{
			UserID:            i,
			CurrentSelections: make([]models.InterestSelection, 10),
			Changes:           make([]InterestChange, 5),
			SessionStart:      time.Now(),
		}
		// Явно игнорируем поля для теста производительности
		_ = session.UserID
		_ = session.CurrentSelections
		_ = session.Changes
		_ = session.SessionStart
	}

	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("Performance test took too long: %v", elapsed)
	}
}

// BenchmarkSessionCreation бенчмарк создания сессий.
func BenchmarkSessionCreation(b *testing.B) {
	for i := range b.N {
		session := &EditSession{
			UserID:            i,
			CurrentSelections: make([]models.InterestSelection, 10),
			Changes:           make([]InterestChange, 5),
			SessionStart:      time.Now(),
		}
		// Явно игнорируем поля для бенчмарка производительности
		_ = session.UserID
		_ = session.CurrentSelections
		_ = session.Changes
		_ = session.SessionStart
	}
}

// BenchmarkStatisticsCalculation бенчмарк расчета статистики.
func BenchmarkStatisticsCalculation(b *testing.B) {
	editor := &IsolatedInterestEditor{}

	session := &EditSession{
		CurrentSelections: make([]models.InterestSelection, 100),
		Changes:           make([]InterestChange, 50),
	}

	// Явно игнорируем поле Changes для бенчмарка
	_ = session.Changes

	b.ResetTimer()

	for range b.N {
		_ = editor.calculateEditStats(session)
	}
}
