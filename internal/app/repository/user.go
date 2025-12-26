package repository

import (
	"metoda/internal/app/ds"
)

// GetUserByID - получение пользователя по ID
func (r *Repository) GetUserByID(id uint) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByLogin - получение пользователя по логину
func (r *Repository) GetUserByLogin(login string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser - создание нового пользователя
func (r *Repository) CreateUser(login, password string, isModerator bool) (*ds.User, error) {
	user := &ds.User{
		Login:       login,
		Password:    password,
		IsModerator: isModerator,
	}
	err := r.db.Create(user).Error
	return user, err
}

// UpdateUser - обновление пользователя
func (r *Repository) UpdateUser(user *ds.User) error {
	return r.db.Save(user).Error
}

