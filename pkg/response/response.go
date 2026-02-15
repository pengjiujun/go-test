package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	// 引入你之前的包
	"test/pkg/util"
)

// Response 标准响应结构体
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"` // 去掉 omitempty，保证 data 始终返回 (哪怕是 null)
}

// ===================================
// 1. 成功返回
// ===================================

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: http.StatusOK,
		Msg:  util.TransBiz(c, "Success", nil),
		Data: data,
	})
}

// ===================================
// 2. 智能错误处理 (核心)
// ===================================

func Fail(c *gin.Context, err error) {

	var msg string

	// --- A. 判断是否为参数校验错误 (Validator) ---
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		// 使用之前的 TransValid 逻辑
		msg = util.TransValid(c, err)

		c.JSON(http.StatusOK, Response{
			Code: http.StatusBadRequest,
			Msg:  msg,
			Data: nil,
		})
		return
	}

	// --- B. 判断是否为 JSON 格式错误 (比如传了字符串给 int 字段) ---
	var unmarshalErr *json.UnmarshalTypeError
	if errors.As(err, &unmarshalErr) {
		c.JSON(http.StatusOK, Response{
			Code: http.StatusPaymentRequired,
			Msg:  util.TransBiz(c, "InvalidJSON", nil), // 需在 TOML 配置
			Data: nil,
		})
		return
	}

	// --- C. 判断是否为自定义业务错误 (BizError) ---
	var bizErr *util.BizError
	if errors.As(err, &bizErr) {
		// 提取 Key 和 Params 进行翻译
		msg = util.TransBiz(c, bizErr.Key, bizErr.Params)

		c.JSON(http.StatusOK, Response{
			Code: http.StatusPaymentRequired,
			Msg:  msg,
			Data: nil,
		})

		return
	}

	// --- D. 其他未知错误 (系统错误) ---
	// 生产环境建议打印日志: log.Println("System Error:", err)
	c.JSON(http.StatusOK, Response{
		Code: http.StatusInternalServerError,
		//Msg:  util.TransBiz(c, "SystemBusy", nil),
		Msg:  util.TransBiz(c, err.Error(), nil),
		Data: nil,
	})
}
