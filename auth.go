package main

import (
	"github.com/alexedwards/argon2id"
	"github.com/benjamineskola/bookmarks/database"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string
	Password string
}

func GetUserByEmail(db *gorm.DB, email string) *User {
	var user User

	db.Where("email = ?", email).First(&user)

	return &user
}

func GetValidatedUser(email string, password string) (*User, error) {
	user := GetUserByEmail(database.DB, email)

	_, err := argon2id.ComparePasswordAndHash(password, user.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}
