package serializer

import (
	"test/internal/model"
	"test/pkg/util"
)

type BannerResp struct {
	ID        uint           `json:"id"`
	ImageUrl  string         `json:"image_url"`
	Sort      int            `json:"sort"`
	UpdatedAt util.LocalTime `json:"updated_at"`
}

func BuildBanner(item model.Banner) BannerResp {
	return BannerResp{
		ID:        item.ID,
		UpdatedAt: util.LocalTime(item.UpdatedAt),
		ImageUrl:  item.ImageUrl,
		Sort:      item.Sort,
	}
}
func BuildBanners(items []model.Banner) []BannerResp {
	var banners []BannerResp
	for _, item := range items {
		banners = append(banners, BuildBanner(item))
	}
	return banners
}
