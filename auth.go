package main

import (
	"fmt"

	"github.com/alexedwards/argon2id"
	"github.com/benjamineskola/bookmarks/database"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string
	Password string
}

func NewUser(email string, password string) (*User, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("could not create password: %w", err)
	}

	user := User{Email: email, Password: hash} //nolint:exhaustruct

	return &user, nil
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
		return nil, fmt.Errorf("could not verify password: %w", err)
	}

	return user, nil
}
