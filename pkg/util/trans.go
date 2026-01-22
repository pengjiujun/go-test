package util

import (
	"strings"

	"github.com/gin-gonic/gin"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// TransBiz 翻译业务消息 (如: 用户已存在)
func TransBiz(c *gin.Context, key string, params map[string]interface{}) string {
	localizer, _ := c.MustGet("localizer").(*i18n.Localizer)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: params,
	})
	if err != nil {
		return key
	}
	return msg
}

// TransValid 翻译校验错误 (核心：包含二次替换逻辑)
func TransValid(c *gin.Context, err error) string {
	vTrans, _ := c.MustGet("vTrans").(ut.Translator)

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			// 1. Validator 翻译: "{Phone}格式不正确" (此时 {Phone} 是占位符)
			msg := e.Translate(vTrans)

			// 2. 提取占位符中的 Key: "{Phone}" -> "Phone"
			// 注意：这里 e.Field() 返回的就是我们在 RegisterTagNameFunc 里写的 "{label}"
			fieldKey := strings.Trim(e.Field(), "{}")

			// 3. 查 TOML 翻译字段名: "Field_Phone" -> "手机号" (或 "Mobile")
			// 这里的 Field_ 前缀是为了和 TOML 里的 Key 对应
			targetField := TransBiz(c, "Field_"+fieldKey, nil)

			// 4. 替换: "{Phone}格式不正确" -> "手机号格式不正确"
			return strings.Replace(msg, e.Field(), targetField, 1)
		}
	}
	return err.Error()
}
