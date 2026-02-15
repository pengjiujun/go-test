package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Account  string  `gorm:"type:varchar(20);not null"`
	Password string  `gorm:"type:varchar(255);not null"`
	Amount   float64 `gorm:"type:decimal(10,2);not null"`
	Nickname string  `gorm:"type:varchar(20);not null"`
}

func (User) TableName() string {
	return "users"
}
