package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Session хранит информацию о сессии пользователя
type Session struct {
	UserID      uint
	Login       string
	IsModerator bool
	ExpiresAt   time.Time
}

// SessionManager управляет сессиями пользователей
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

var (
	sessionManager *SessionManager
	sessionOnce   sync.Once
)

// GetSessionManager возвращает singleton экземпляр SessionManager
func GetSessionManager() *SessionManager {
	sessionOnce.Do(func() {
		sessionManager = &SessionManager{
			sessions: make(map[string]*Session),
		}
		// Запускаем очистку истекших сессий
		go sessionManager.cleanupExpiredSessions()
	})
	return sessionManager
}

// CreateSession создает новую сессию для пользователя
func (sm *SessionManager) CreateSession(userID uint, login string, isModerator bool) string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Генерируем уникальный токен сессии
	token := generateSessionToken()
	
	sm.sessions[token] = &Session{
		UserID:      userID,
		Login:       login,
		IsModerator: isModerator,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Сессия на 24 часа
	}

	return token
}

// GetSession возвращает сессию по токену
func (sm *SessionManager) GetSession(token string) (*Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return nil, false
	}

	// Проверяем, не истекла ли сессия
	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, token)
		return nil, false
	}

	return session, true
}

// DeleteSession удаляет сессию
func (sm *SessionManager) DeleteSession(token string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, token)
}

// cleanupExpiredSessions периодически очищает истекшие сессии
func (sm *SessionManager) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for token, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, token)
			}
		}
		sm.mu.Unlock()
	}
}

// generateSessionToken генерирует случайный токен для сессии
func generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

