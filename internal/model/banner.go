package model

import "gorm.io/gorm"

type Banner struct {
	gorm.Model
	ImageUrl string `gorm:"type:varchar(255);"`
	Sort     int    `gorm:"type:int;"`
}

func (Banner) TableName() string {
	return "banners"
}
