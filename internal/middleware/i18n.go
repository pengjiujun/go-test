package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"

	// 引入你自己的 translation 包
	// 假设你的 go.mod 第一行写的是 module example
	"test/pkg/translation"
)

func I18nMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取 Header
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = "zh" // 默认中文
		}

		// 2. 准备 Localizer (业务翻译用)
		localizer := i18n.NewLocalizer(translation.I18nBundle, lang)
		c.Set("localizer", localizer)

		// 3. 准备 Validator Translator
		vTrans, found := translation.Uni.FindTranslator(lang)
		if !found {
			vTrans, _ = translation.Uni.GetTranslator("zh") // 找不到日语才兜底回中文
		}
		c.Set("vTrans", vTrans)
		c.Next()
	}
}
