package service

import (
	"context"
	"test/internal/dao"
	"test/internal/model"
	"test/pkg/database"
	"test/pkg/util"
)

func ListBanners(ctx context.Context) ([]model.Banner, error) {
	bannerDao := dao.NewBannerDao(database.DB)
	data, err := bannerDao.ListBanner(ctx)
	if err != nil {
		return nil, util.NewBizErr("SystemBusy", nil)
	}
	return *data, nil
}
