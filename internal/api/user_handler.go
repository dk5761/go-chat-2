package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	apperrors "github.com/chat-backend/internal/apperrors"
	"github.com/chat-backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var input service.RegisterUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				field := e.Field()
				switch field {
				case "Username":
					errorMessages[field] = "Username is required"
				case "Email":
					if e.Tag() == "email" {
						errorMessages[field] = "Please enter a valid email address"
					} else {
						errorMessages[field] = "Email is required"
					}
				case "Password":
					if e.Tag() == "min" {
						errorMessages[field] = "Password must be at least 8 characters long"
					} else {
						errorMessages[field] = "Password is required"
					}
				case "FullName":
					errorMessages[field] = "Full name is required"
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	response, err := h.userService.Register(c.Request.Context(), input)
	if err != nil {
		switch err {
		case apperrors.ErrEmailExists, apperrors.ErrUsernameExists:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case apperrors.ErrServerError:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, response)
}

func (h *UserHandler) Login(c *gin.Context) {
	var input service.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				field := e.Field()
				switch field {
				case "Email":
					if e.Tag() == "email" {
						errorMessages[field] = "Please enter a valid email address"
					} else {
						errorMessages[field] = "Email is required"
					}
				case "Password":
					errorMessages[field] = "Password is required"
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	response, err := h.userService.Login(c.Request.Context(), input)
	if err != nil {
		switch err {
		case apperrors.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case apperrors.ErrServerError:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidUserID.Error()})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		switch err {
		case apperrors.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case apperrors.ErrServerError:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdatePassword(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": apperrors.ErrInvalidUserID.Error()})
		return
	}

	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errorMessages := make(map[string]string)
			for _, e := range validationErrors {
				field := e.Field()
				switch field {
				case "OldPassword":
					errorMessages[field] = "Current password is required"
				case "NewPassword":
					if e.Tag() == "min" {
						errorMessages[field] = "New password must be at least 8 characters long"
					} else {
						errorMessages[field] = "New password is required"
					}
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMessages})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input format"})
		return
	}

	if err := h.userService.UpdatePassword(c.Request.Context(), userID, input.OldPassword, input.NewPassword); err != nil {
		switch err {
		case apperrors.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case apperrors.ErrWeakPassword:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case apperrors.ErrServerError:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (h *UserHandler) GetUserStatus(c *gin.Context) {
	userID := c.Param("id")

	status, err := h.userService.GetUserStatus(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

func (h *UserHandler) GetMultiUserStatus(c *gin.Context) {
	var input struct {
		UserIDs []string `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	statuses, err := h.userService.GetMultiUserStatus(c.Request.Context(), input.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

// RegisterRoutes registers the user routes
func (h *UserHandler) RegisterRoutes(router *gin.RouterGroup) {
	users := router.Group("/users")
	{
		users.POST("/register", h.Register)
		users.POST("/login", h.Login)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id/password", h.UpdatePassword)
		users.GET("/:id/status", h.GetUserStatus)
		users.POST("/status/multi", h.GetMultiUserStatus)
	}
}
