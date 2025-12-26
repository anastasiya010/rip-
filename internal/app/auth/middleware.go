package auth

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	SessionCookieName = "session_token"
	UserIDKey         = "user_id"
	UserLoginKey      = "user_login"
	IsModeratorKey    = "is_moderator"
)

// AuthMiddleware проверяет авторизацию пользователя
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен сессии из cookie
		token, err := c.Cookie(SessionCookieName)
		if err != nil || token == "" {
			// Нет сессии - пользователь гость
			logrus.Debugf("AuthMiddleware: No cookie found - path=%s, error=%v", c.Request.URL.Path, err)
			c.Set("user_role", "guest")
			c.Next()
			return
		}

		// Декодируем URL-кодированное значение cookie (Postman может кодировать специальные символы)
		decodedToken, decodeErr := url.QueryUnescape(token)
		if decodeErr == nil && decodedToken != token {
			logrus.Debugf("AuthMiddleware: Cookie URL-decoded")
			token = decodedToken
		}

		logrus.Debugf("AuthMiddleware: Cookie found - token=%s (first 20 chars)", token[:min(20, len(token))])

		// Проверяем сессию
		sm := GetSessionManager()
		session, exists := sm.GetSession(token)
		if !exists {
			// Сессия недействительна - удаляем cookie
			logrus.Warnf("AuthMiddleware: Session not found for token (first 20 chars)=%s", token[:min(20, len(token))])
			c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
			c.Set("user_role", "guest")
			c.Next()
			return
		}

		logrus.Infof("AuthMiddleware: Session found - UserID=%d, Login=%s, IsModerator=%v, Path=%s", 
			session.UserID, session.Login, session.IsModerator, c.Request.URL.Path)

		// Устанавливаем данные пользователя в контекст
		c.Set(UserIDKey, session.UserID)
		c.Set(UserLoginKey, session.Login)
		c.Set(IsModeratorKey, session.IsModerator)
		
		if session.IsModerator {
			c.Set("user_role", "admin")
		} else {
			c.Set("user_role", "user")
		}

		c.Next()
	}
}

// RequireAuth требует авторизации (не гость)
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole == "guest" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "fail",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin требует прав администратора
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists || userRole != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "fail",
				"message": "Admin access required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID получает ID пользователя из контекста
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUserLogin получает логин пользователя из контекста
func GetUserLogin(c *gin.Context) (string, bool) {
	login, exists := c.Get(UserLoginKey)
	if !exists {
		return "", false
	}
	loginStr, ok := login.(string)
	return loginStr, ok
}

// IsModerator проверяет, является ли пользователь модератором
func IsModerator(c *gin.Context) bool {
	isMod, exists := c.Get(IsModeratorKey)
	if !exists {
		logrus.Debugf("IsModerator: Key 'is_moderator' not found in context for path=%s", c.Request.URL.Path)
		return false
	}
	isModBool, ok := isMod.(bool)
	if !ok {
		logrus.Warnf("IsModerator: Key 'is_moderator' exists but type is %T, value=%v for path=%s", isMod, isMod, c.Request.URL.Path)
		return false
	}
	logrus.Debugf("IsModerator: Returning %v for path=%s", isModBool, c.Request.URL.Path)
	return isModBool
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

