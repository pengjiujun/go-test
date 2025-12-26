package handle

import "net/http"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func SuccessResponse(data interface{}) *Response {
	return &Response{
		Code: http.StatusOK,
		Msg:  "success",
		Data: data,
	}
}

func ErrorResponse(code int, msg string) *Response {
	return &Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	}
}
