package model

import "gorm.io/gorm"

type Address struct {
	gorm.Model
}

func (Address) TableName() string {
	return "banners"
}
