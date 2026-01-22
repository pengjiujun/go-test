package request

type RegisterReq struct {
	// label: 对应 TOML 里的 Field_Phone
	Account         string `json:"account" form:"account" binding:"required,mobile" label:"Phone"`
	Password        string `json:"password" form:"password" binding:"required" label:"Password"`
	PasswordConfirm string `json:"password_confirm" form:"password_confirm" binding:"required,eqfield=Password" label:"PasswordConfirm"`
}

// LoginReq 登录请求参数
type LoginReq struct {
	Account  string `json:"account" form:"account" binding:"required" label:"Phone"`
	Password string `json:"password" form:"password" binding:"required" label:"Password"`
}
