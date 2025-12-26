package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Account  string `gorm:"type:varchar(20);not null"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string {
	return "users"
}
