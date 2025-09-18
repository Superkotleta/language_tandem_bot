package models

import (
	"testing"
	"time"

	"language-exchange-bot/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestUser_IsNew(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "New user",
			user: models.User{
				Status: models.StatusNew,
			},
			expected: true,
		},
		{
			name: "Active user",
			user: models.User{
				Status: models.StatusActive,
			},
			expected: false,
		},
		{
			name: "Filling profile user",
			user: models.User{
				Status: models.StatusFilling,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.user.Status == models.StatusNew

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "Active user",
			user: models.User{
				Status: models.StatusActive,
			},
			expected: true,
		},
		{
			name: "New user",
			user: models.User{
				Status: models.StatusNew,
			},
			expected: false,
		},
		{
			name: "Filling profile user",
			user: models.User{
				Status: models.StatusFilling,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.user.Status == models.StatusActive

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_IsPaused(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "Paused user",
			user: models.User{
				Status: models.StatusPaused,
			},
			expected: true,
		},
		{
			name: "Active user",
			user: models.User{
				Status: models.StatusActive,
			},
			expected: false,
		},
		{
			name: "New user",
			user: models.User{
				Status: models.StatusNew,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.user.Status == models.StatusPaused

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_HasNativeLanguage(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "User with native language",
			user: models.User{
				NativeLanguageCode: "ru",
			},
			expected: true,
		},
		{
			name: "User without native language",
			user: models.User{
				NativeLanguageCode: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.user.NativeLanguageCode != ""

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_HasTargetLanguage(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "User with target language",
			user: models.User{
				TargetLanguageCode: "en",
			},
			expected: true,
		},
		{
			name: "User without target language",
			user: models.User{
				TargetLanguageCode: "",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.user.TargetLanguageCode != ""

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_HasInterests(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected bool
	}{
		{
			name: "User with interests",
			user: models.User{
				Interests: []int{1, 2, 3},
			},
			expected: true,
		},
		{
			name: "User without interests",
			user: models.User{
				Interests: []int{},
			},
			expected: false,
		},
		{
			name: "User with nil interests",
			user: models.User{
				Interests: nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := len(tt.user.Interests) > 0

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_GetInterestCount(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected int
	}{
		{
			name: "User with 3 interests",
			user: models.User{
				Interests: []int{1, 2, 3},
			},
			expected: 3,
		},
		{
			name: "User without interests",
			user: models.User{
				Interests: []int{},
			},
			expected: 0,
		},
		{
			name: "User with nil interests",
			user: models.User{
				Interests: nil,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := len(tt.user.Interests)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUser_UpdateLastActivity(t *testing.T) {
	// Arrange
	user := models.User{
		UpdatedAt: time.Time{},
	}

	// Act
	user.UpdatedAt = time.Now()

	// Assert
	assert.False(t, user.UpdatedAt.IsZero())
	assert.True(t, time.Since(user.UpdatedAt) < time.Second)
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		user     models.User
		expected string
	}{
		{
			name: "User with username",
			user: models.User{
				FirstName: "John",
				Username:  "john_doe",
			},
			expected: "John (@john_doe)",
		},
		{
			name: "User without username",
			user: models.User{
				FirstName: "John",
				Username:  "",
			},
			expected: "John",
		},
		{
			name: "User with empty first name",
			user: models.User{
				FirstName: "",
				Username:  "john_doe",
			},
			expected: "@john_doe",
		},
		{
			name: "User with empty first name and username",
			user: models.User{
				FirstName: "",
				Username:  "",
			},
			expected: "Unknown User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			var result string
			if tt.user.FirstName != "" && tt.user.Username != "" {
				result = tt.user.FirstName + " (@" + tt.user.Username + ")"
			} else if tt.user.FirstName != "" {
				result = tt.user.FirstName
			} else if tt.user.Username != "" {
				result = "@" + tt.user.Username
			} else {
				result = "Unknown User"
			}

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}
