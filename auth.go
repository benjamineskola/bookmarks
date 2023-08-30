package main

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Email    string
	Password string
}

func GetPasswordForUser(db *gorm.DB, email string) string {
	var user User

	db.Where("email = ?", email).First(&user)

	return user.Password
}
