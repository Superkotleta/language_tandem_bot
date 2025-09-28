package handlers

import (
	"net/http"
	"strconv"

	"profile/internal/models"
	"profile/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ProfileHandler struct {
	profileService *service.ProfileService
	validator      *validator.Validate
}

func NewProfileHandler(profileService *service.ProfileService, validator *validator.Validate) *ProfileHandler {
	return &ProfileHandler{
		profileService: profileService,
		validator:      validator,
	}
}

// @Router /users [post].
func (h *ProfileHandler) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	// Create user
	user, err := h.profileService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// @Router /users/{id} [get].
func (h *ProfileHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid integer",
		})
		return
	}

	user, err := h.profileService.GetUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Router /users/telegram/{telegram_id} [get].
func (h *ProfileHandler) GetUserByTelegramID(c *gin.Context) {
	telegramIDStr := c.Param("telegram_id")
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid Telegram ID",
			Message: "Telegram ID must be a valid integer",
		})
		return
	}

	user, err := h.profileService.GetUserByTelegramID(c.Request.Context(), telegramID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Router /users/discord/{discord_id} [get].
func (h *ProfileHandler) GetUserByDiscordID(c *gin.Context) {
	discordIDStr := c.Param("discord_id")
	discordID, err := strconv.ParseInt(discordIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid Discord ID",
			Message: "Discord ID must be a valid integer",
		})
		return
	}

	user, err := h.profileService.GetUserByDiscordID(c.Request.Context(), discordID)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Router /users/{id} [put].
func (h *ProfileHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid integer",
		})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	user, err := h.profileService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// @Router /users/{id} [delete].
func (h *ProfileHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid integer",
		})
		return
	}

	err = h.profileService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete user",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Router /users [get].
func (h *ProfileHandler) ListUsers(c *gin.Context) {
	var req models.UserSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid query parameters",
			Message: err.Error(),
		})
		return
	}

	// Validate request
	if err := h.validator.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	users, err := h.profileService.ListUsers(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to list users",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, users)
}

// @Router /users/{id}/last-seen [put].
func (h *ProfileHandler) UpdateLastSeen(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid integer",
		})
		return
	}

	err = h.profileService.UpdateLastSeen(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update last seen",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Router /users/{id}/completion [get].
func (h *ProfileHandler) GetUserProfileCompletion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid user ID",
			Message: "User ID must be a valid integer",
		})
		return
	}

	score, err := h.profileService.GetUserProfileCompletion(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "User not found",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get profile completion",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"completion_score": score,
	})
}
