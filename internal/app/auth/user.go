package auth

import (
	"sync"
)

var (
	// Фиксированный пользователь-создатель для ЛР3
	fixedCreatorID uint = 1
	once           sync.Once
	instance       *UserService
)

// UserService - singleton для работы с пользователем
type UserService struct {
	creatorID uint
}

// GetUserService возвращает singleton экземпляр UserService
func GetUserService() *UserService {
	once.Do(func() {
		instance = &UserService{
			creatorID: fixedCreatorID,
		}
	})
	return instance
}

// GetCreatorID возвращает ID фиксированного пользователя-создателя
func (us *UserService) GetCreatorID() uint {
	return us.creatorID
}

// SetCreatorID устанавливает ID пользователя-создателя (для тестирования)
func (us *UserService) SetCreatorID(id uint) {
	us.creatorID = id
}

