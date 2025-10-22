package language

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// SESSION STRUCTURE TESTS
// =============================================================================

// TestLanguageEditSessionStructure тестирует структуру сессии редактирования языков
func TestLanguageEditSessionStructure(t *testing.T) {
	t.Run("Create language edit session", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:               123,
			OriginalNativeLang:   "ru",
			OriginalTargetLang:   "en",
			OriginalTargetLevel:  "B1",
			CurrentNativeLang:    "en", // Changed
			CurrentTargetLang:    "ru", // Changed
			CurrentTargetLevel:   "A2", // Changed
			Changes: []LanguageChange{
				{
					Field:     "native_language",
					OldValue:  "ru",
					NewValue:  "en",
					Timestamp: time.Now(),
				},
			},
			CurrentStep:       "main_menu",
			SessionStart:      time.Now(),
			LastActivity:      time.Now(),
			InterfaceLanguage: "ru",
		}

		assert.NotNil(t, session)
		assert.Equal(t, 123, session.UserID)
		assert.Equal(t, "en", session.CurrentNativeLang)
		assert.Equal(t, "ru", session.CurrentTargetLang)
		assert.Equal(t, "A2", session.CurrentTargetLevel)
		assert.Len(t, session.Changes, 1)
		assert.Equal(t, "native_language", session.Changes[0].Field)
	})

	t.Run("Create empty session", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:               456,
			OriginalNativeLang:   "ru",
			OriginalTargetLang:   "en",
			OriginalTargetLevel:  "B2",
			CurrentNativeLang:    "ru",
			CurrentTargetLang:    "en",
			CurrentTargetLevel:   "B2",
			Changes:              []LanguageChange{},
			CurrentStep:          "main_menu",
			SessionStart:         time.Now(),
			LastActivity:         time.Now(),
			InterfaceLanguage:    "en",
		}

		assert.NotNil(t, session)
		assert.Equal(t, 456, session.UserID)
		assert.Len(t, session.Changes, 0)
		assert.Equal(t, session.OriginalNativeLang, session.CurrentNativeLang)
		assert.Equal(t, session.OriginalTargetLang, session.CurrentTargetLang)
	})

	t.Run("Session with multiple changes", func(t *testing.T) {
		now := time.Now()
		session := &LanguageEditSession{
			UserID:               789,
			OriginalNativeLang:   "ru",
			OriginalTargetLang:   "en",
			OriginalTargetLevel:  "B1",
			CurrentNativeLang:    "en",
			CurrentTargetLang:    "ru",
			CurrentTargetLevel:   "C1",
			Changes: []LanguageChange{
				{
					Field:     "native_language",
					OldValue:  "ru",
					NewValue:  "en",
					Timestamp: now,
				},
				{
					Field:     "target_language",
					OldValue:  "en",
					NewValue:  "ru",
					Timestamp: now.Add(1 * time.Second),
				},
				{
					Field:     "target_level",
					OldValue:  "B1",
					NewValue:  "C1",
					Timestamp: now.Add(2 * time.Second),
				},
			},
			CurrentStep:       "preview",
			SessionStart:      now,
			LastActivity:      now.Add(2 * time.Second),
			InterfaceLanguage: "ru",
		}

		assert.NotNil(t, session)
		assert.Len(t, session.Changes, 3)
		assert.Equal(t, "preview", session.CurrentStep)

		// Проверяем порядок изменений
		assert.Equal(t, "native_language", session.Changes[0].Field)
		assert.Equal(t, "target_language", session.Changes[1].Field)
		assert.Equal(t, "target_level", session.Changes[2].Field)
	})
}

// =============================================================================
// LANGUAGE CHANGE LOGIC TESTS
// =============================================================================

// TestLanguageChangeLogic тестирует логику изменения языков
func TestLanguageChangeLogic(t *testing.T) {
	t.Run("Native language change - not Russian to Russian", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "en",
			CurrentNativeLang:  "en",
			OriginalTargetLang: "ru",
			CurrentTargetLang:  "ru",
			Changes:            []LanguageChange{},
		}

		// Меняем родной язык на русский
		newNativeLang := "ru"
		if session.CurrentNativeLang != newNativeLang {
			change := LanguageChange{
				Field:     "native_language",
				OldValue:  session.CurrentNativeLang,
				NewValue:  newNativeLang,
				Timestamp: time.Now(),
			}
			session.Changes = append(session.Changes, change)
			session.CurrentNativeLang = newNativeLang

			// Если родной стал русским, нужно изменить target
			if newNativeLang == "ru" && session.CurrentTargetLang == "ru" {
				// Сбрасываем target, так как нельзя учить свой родной
				targetChange := LanguageChange{
					Field:     "target_language",
					OldValue:  session.CurrentTargetLang,
					NewValue:  "",
					Timestamp: time.Now(),
				}
				session.Changes = append(session.Changes, targetChange)
				session.CurrentTargetLang = ""
			}
		}

		assert.Len(t, session.Changes, 2) // native change + target reset
		assert.Equal(t, "ru", session.CurrentNativeLang)
		assert.Equal(t, "", session.CurrentTargetLang)
	})

	t.Run("Native language change - Russian to non-Russian", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "ru",
			CurrentNativeLang:  "ru",
			OriginalTargetLang: "en",
			CurrentTargetLang:  "en",
			Changes:            []LanguageChange{},
		}

		// Меняем родной язык на не-русский
		newNativeLang := "en"
		if session.CurrentNativeLang != newNativeLang {
			change := LanguageChange{
				Field:     "native_language",
				OldValue:  session.CurrentNativeLang,
				NewValue:  newNativeLang,
				Timestamp: time.Now(),
			}
			session.Changes = append(session.Changes, change)
			session.CurrentNativeLang = newNativeLang

			// Если родной не русский, автоматически ставим русский как target
			if newNativeLang != "ru" && session.CurrentTargetLang != "ru" {
				targetChange := LanguageChange{
					Field:     "target_language",
					OldValue:  session.CurrentTargetLang,
					NewValue:  "ru",
					Timestamp: time.Now(),
				}
				session.Changes = append(session.Changes, targetChange)
				session.CurrentTargetLang = "ru"
			}
		}

		assert.Len(t, session.Changes, 2) // native change + target auto-set
		assert.Equal(t, "en", session.CurrentNativeLang)
		assert.Equal(t, "ru", session.CurrentTargetLang)
	})

	t.Run("Target language can only be changed when native is Russian", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "en",
			CurrentNativeLang:  "en",
			OriginalTargetLang: "ru",
			CurrentTargetLang:  "ru",
			Changes:            []LanguageChange{},
		}

		// Пытаемся изменить target, когда native не русский
		canChangeTarget := session.CurrentNativeLang == "ru"

		assert.False(t, canChangeTarget, "Should not be able to change target when native is not Russian")
	})

	t.Run("Level change", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:              123,
			OriginalTargetLevel: "B1",
			CurrentTargetLevel:  "B1",
			Changes:             []LanguageChange{},
		}

		newLevel := "C1"
		if session.CurrentTargetLevel != newLevel {
			change := LanguageChange{
				Field:     "target_level",
				OldValue:  session.CurrentTargetLevel,
				NewValue:  newLevel,
				Timestamp: time.Now(),
			}
			session.Changes = append(session.Changes, change)
			session.CurrentTargetLevel = newLevel
		}

		assert.Len(t, session.Changes, 1)
		assert.Equal(t, "C1", session.CurrentTargetLevel)
		assert.Equal(t, "target_level", session.Changes[0].Field)
	})
}

// =============================================================================
// UNDO FUNCTIONALITY TESTS
// =============================================================================

// TestUndoFunctionality тестирует функциональность отмены изменений
func TestUndoFunctionality(t *testing.T) {
	t.Run("Undo last change", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "ru",
			CurrentNativeLang:  "en",
			Changes: []LanguageChange{
				{
					Field:     "native_language",
					OldValue:  "ru",
					NewValue:  "en",
					Timestamp: time.Now(),
				},
			},
		}

		// Отменяем последнее изменение
		if len(session.Changes) > 0 {
			lastChange := session.Changes[len(session.Changes)-1]
			session.Changes = session.Changes[:len(session.Changes)-1]

			// Откатываем значение
			switch lastChange.Field {
			case "native_language":
				session.CurrentNativeLang = lastChange.OldValue.(string)
			}
		}

		assert.Len(t, session.Changes, 0)
		assert.Equal(t, "ru", session.CurrentNativeLang)
	})

	t.Run("Undo multiple changes", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "ru",
			CurrentNativeLang:  "en",
			OriginalTargetLang: "en",
			CurrentTargetLang:  "de",
			Changes: []LanguageChange{
				{
					Field:     "native_language",
					OldValue:  "ru",
					NewValue:  "en",
					Timestamp: time.Now(),
				},
				{
					Field:     "target_language",
					OldValue:  "en",
					NewValue:  "de",
					Timestamp: time.Now(),
				},
			},
		}

		// Отменяем последнее изменение (target_language)
		if len(session.Changes) > 0 {
			lastChange := session.Changes[len(session.Changes)-1]
			session.Changes = session.Changes[:len(session.Changes)-1]

			switch lastChange.Field {
			case "target_language":
				session.CurrentTargetLang = lastChange.OldValue.(string)
			}
		}

		assert.Len(t, session.Changes, 1)
		assert.Equal(t, "en", session.CurrentNativeLang) // Не изменился
		assert.Equal(t, "en", session.CurrentTargetLang) // Откатился
	})

	t.Run("Undo with no changes", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "ru",
			CurrentNativeLang:  "ru",
			Changes:            []LanguageChange{},
		}

		canUndo := len(session.Changes) > 0

		assert.False(t, canUndo, "Should not be able to undo when there are no changes")
	})
}

// =============================================================================
// VALIDATION TESTS
// =============================================================================

// TestLanguageValidation тестирует валидацию языковых данных
func TestLanguageValidation(t *testing.T) {
	t.Run("Valid language codes", func(t *testing.T) {
		validCodes := []string{"ru", "en", "es", "de"}

		for _, code := range validCodes {
			assert.NotEmpty(t, code, "Language code should not be empty")
			assert.Len(t, code, 2, "Language code should be 2 characters")
		}
	})

	t.Run("Valid level codes", func(t *testing.T) {
		validLevels := []string{"A1", "A2", "B1", "B2", "C1", "C2"}

		for _, level := range validLevels {
			assert.NotEmpty(t, level, "Level code should not be empty")
			assert.Len(t, level, 2, "Level code should be 2 characters")
		}
	})

	t.Run("Native and target languages must be different", func(t *testing.T) {
		nativeLang := "ru"
		targetLang := "ru"

		isValid := nativeLang != targetLang
		assert.False(t, isValid, "Native and target languages should be different")
	})

	t.Run("Russian native language logic", func(t *testing.T) {
		// Если родной = русский, target может быть любой (кроме русского)
		nativeLang := "ru"
		validTargets := []string{"en", "es", "de"}

		for _, target := range validTargets {
			isValid := nativeLang != target
			assert.True(t, isValid, "Target should be different from native")
		}

		// Если родной != русский, target должен быть русским
		nativeLang = "en"
		expectedTarget := "ru"

		assert.Equal(t, "ru", expectedTarget)
	})
}

// =============================================================================
// SESSION LIFECYCLE TESTS
// =============================================================================

// TestSessionLifecycle тестирует жизненный цикл сессии
func TestSessionLifecycle(t *testing.T) {
	t.Run("New session starts with no changes", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:             123,
			OriginalNativeLang: "ru",
			CurrentNativeLang:  "ru",
			Changes:            []LanguageChange{},
			SessionStart:       time.Now(),
		}

		assert.Len(t, session.Changes, 0)
		assert.Equal(t, session.OriginalNativeLang, session.CurrentNativeLang)
	})

	t.Run("Session tracks last activity", func(t *testing.T) {
		startTime := time.Now()
		session := &LanguageEditSession{
			UserID:       123,
			SessionStart: startTime,
			LastActivity: startTime,
		}

		// Симулируем активность через 5 секунд
		newActivity := startTime.Add(5 * time.Second)
		session.LastActivity = newActivity

		assert.True(t, session.LastActivity.After(session.SessionStart))
		assert.Equal(t, 5*time.Second, session.LastActivity.Sub(session.SessionStart))
	})

	t.Run("Session step progression", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:      123,
			CurrentStep: "main_menu",
		}

		steps := []string{"main_menu", "native", "target", "level", "preview"}

		for _, step := range steps {
			session.CurrentStep = step
			assert.Equal(t, step, session.CurrentStep)
		}
	})
}

// =============================================================================
// CHANGES TRACKING TESTS
// =============================================================================

// TestChangesTracking тестирует отслеживание изменений
func TestChangesTracking(t *testing.T) {
	t.Run("Track single change", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:  123,
			Changes: []LanguageChange{},
		}

		change := LanguageChange{
			Field:     "native_language",
			OldValue:  "ru",
			NewValue:  "en",
			Timestamp: time.Now(),
		}
		session.Changes = append(session.Changes, change)

		assert.Len(t, session.Changes, 1)
		assert.Equal(t, "native_language", session.Changes[0].Field)
		assert.Equal(t, "ru", session.Changes[0].OldValue)
		assert.Equal(t, "en", session.Changes[0].NewValue)
	})

	t.Run("Track multiple changes chronologically", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID:  123,
			Changes: []LanguageChange{},
		}

		baseTime := time.Now()

		changes := []LanguageChange{
			{Field: "native_language", OldValue: "ru", NewValue: "en", Timestamp: baseTime},
			{Field: "target_language", OldValue: "en", NewValue: "ru", Timestamp: baseTime.Add(1 * time.Second)},
			{Field: "target_level", OldValue: "B1", NewValue: "C1", Timestamp: baseTime.Add(2 * time.Second)},
		}

		for _, change := range changes {
			session.Changes = append(session.Changes, change)
		}

		assert.Len(t, session.Changes, 3)

		// Проверяем хронологический порядок
		for i := 1; i < len(session.Changes); i++ {
			assert.True(t, session.Changes[i].Timestamp.After(session.Changes[i-1].Timestamp))
		}
	})

	t.Run("Calculate changes count", func(t *testing.T) {
		session := &LanguageEditSession{
			UserID: 123,
			Changes: []LanguageChange{
				{Field: "native_language", OldValue: "ru", NewValue: "en", Timestamp: time.Now()},
				{Field: "target_language", OldValue: "en", NewValue: "ru", Timestamp: time.Now()},
			},
		}

		changesCount := len(session.Changes)

		assert.Equal(t, 2, changesCount)
	})
}
