package util

// BizError 自定义业务错误
type BizError struct {
	Key    string                 // TOML 里的 Key
	Params map[string]interface{} // 动态参数
}

// 实现 error 接口，这样它就可以当做 error 返回
func (e *BizError) Error() string {
	return e.Key
}

// NewBizErr 快速创建业务错误的辅助函数
func NewBizErr(key string, params map[string]interface{}) *BizError {
	return &BizError{
		Key:    key,
		Params: params,
	}
}
