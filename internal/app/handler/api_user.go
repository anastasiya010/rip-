package handler

import (
	"net/http"
	"time"

	"metoda/internal/app/auth"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterUserAPI - POST /api/users/register - регистрация
func (h *Handler) RegisterUserAPI(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	// Проверяем, существует ли пользователь
	existing, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("Error checking user existence: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	if existing != nil {
		h.errorResponse(ctx, http.StatusConflict, "User with this login already exists")
		return
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.Errorf("Error hashing password: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Создаем пользователя (только с ролью пользователя, не админа)
	user, err := h.Repository.CreateUser(req.Login, string(hashedPassword), false)
	if err != nil {
		logrus.Errorf("Error creating user: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Создаем сессию для нового пользователя
	sm := auth.GetSessionManager()
	sessionToken := sm.CreateSession(user.ID, user.Login, user.IsModerator)

	// Устанавливаем cookie с сессией
	// Для localhost используем "localhost" как domain, чтобы Postman правильно сохранял cookie
	ctx.SetCookie(auth.SessionCookieName, sessionToken, int(24*time.Hour.Seconds()), "/", "localhost", false, true)

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

// GetUserProfileAPI - GET /api/users/profile - поля пользователя после аутентификации
func (h *Handler) GetUserProfileAPI(ctx *gin.Context) {
	// Получаем ID пользователя из сессии
	userID, exists := auth.GetUserID(ctx)
	if !exists {
		h.errorResponse(ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "User not found")
			return
		}
		logrus.Errorf("Error getting user: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

// UpdateUserProfileAPI - PUT /api/users/profile - обновление пользователя (личный кабинет)
func (h *Handler) UpdateUserProfileAPI(ctx *gin.Context) {
	// Получаем ID пользователя из сессии
	userID, exists := auth.GetUserID(ctx)
	if !exists {
		h.errorResponse(ctx, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusNotFound, "User not found")
			return
		}
		logrus.Errorf("Error getting user: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Обновляем логин если передан
	if req.Login != "" {
		// Проверяем, не занят ли логин другим пользователем
		existing, err := h.Repository.GetUserByLogin(req.Login)
		if err != nil && err != gorm.ErrRecordNotFound {
			logrus.Errorf("Error checking user existence: %v", err)
			h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
			return
		}
		if existing != nil && existing.ID != userID {
			h.errorResponse(ctx, http.StatusConflict, "Login already taken")
			return
		}
		user.Login = req.Login
	}

	// Обновляем пароль если передан
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			logrus.Errorf("Error hashing password: %v", err)
			h.errorResponse(ctx, http.StatusInternalServerError, "Failed to process password")
			return
		}
		user.Password = string(hashedPassword)
	}

	err = h.Repository.UpdateUser(user)
	if err != nil {
		logrus.Errorf("Error updating user: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Failed to update user")
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "User updated successfully",
		"data": map[string]interface{}{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

// AuthenticateUserAPI - POST /api/users/auth - аутентификация
func (h *Handler) AuthenticateUserAPI(ctx *gin.Context) {
	var req struct {
		Login    string `json:"login" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorResponse(ctx, http.StatusBadRequest, "Invalid request data: "+err.Error())
		return
	}

	user, err := h.Repository.GetUserByLogin(req.Login)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.errorResponse(ctx, http.StatusUnauthorized, "Invalid login or password")
			return
		}
		logrus.Errorf("Error getting user: %v", err)
		h.errorResponse(ctx, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		h.errorResponse(ctx, http.StatusUnauthorized, "Invalid login or password")
		return
	}

	// Создаем сессию для пользователя
	sm := auth.GetSessionManager()
	sessionToken := sm.CreateSession(user.ID, user.Login, user.IsModerator)

	// Устанавливаем cookie с сессией
	// Для localhost используем "localhost" как domain, чтобы Postman правильно сохранял cookie
	ctx.SetCookie(auth.SessionCookieName, sessionToken, int(24*time.Hour.Seconds()), "/", "localhost", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": map[string]interface{}{
			"id":           user.ID,
			"login":        user.Login,
			"is_moderator": user.IsModerator,
		},
	})
}

// LogoutUserAPI - POST /api/users/logout - деавторизация
func (h *Handler) LogoutUserAPI(ctx *gin.Context) {
	// Получаем токен сессии из cookie
	token, err := ctx.Cookie(auth.SessionCookieName)
	if err == nil && token != "" {
		// Удаляем сессию
		sm := auth.GetSessionManager()
		sm.DeleteSession(token)
	}

	// Удаляем cookie
	ctx.SetCookie(auth.SessionCookieName, "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Logged out successfully",
	})
}
