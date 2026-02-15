package controller

import (
	"github.com/gin-gonic/gin"
	"test/internal/serializer"
	"test/internal/service"
	"test/pkg/response"
)

type BannerController struct{}

func NewBannerController() *BannerController {
	return &BannerController{}
}

func (banner BannerController) Index(c *gin.Context) {
	list, err := service.ListBanners(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, serializer.BuildBanners(list))
}
