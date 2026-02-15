package dao

import (
	"context"
	"gorm.io/gorm"
	"test/internal/model"
)

type BannerDao struct {
	db *gorm.DB
}

func NewBannerDao(db *gorm.DB) *BannerDao {
	return &BannerDao{db: db}
}

func (dao *BannerDao) ListBanner(ctx context.Context) (*[]model.Banner, error) {

	var banner []model.Banner
	err := dao.db.WithContext(ctx).Order("sort desc").Find(&banner).Error
	if err != nil {
		return nil, err
	}
	return &banner, nil
}
