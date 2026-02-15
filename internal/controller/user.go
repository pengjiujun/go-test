package controller

import (
	"github.com/gin-gonic/gin"
	"test/internal/request"
	"test/internal/serializer"
	"test/internal/service"
	"test/pkg/response"
	"test/pkg/util"
)

type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

// Index 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表数据
// @Tags User
// @Accept  json
// @Produce  json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// 👇 修改这一行：data 直接指向 UserDataList
// @Success 200 {object} response.Response{data=serializer.UserDataList} "成功返回"
// @Failure 500 {object} response.Response "系统繁忙"
// @Router /users [get]
func (u *UserController) Index(c *gin.Context) {
	// 1. 绑定分页参数 (自动解析 ?page=1&page_size=10)
	var p util.PaginationReq
	if err := c.ShouldBindQuery(&p); err != nil {
		// 如果参数格式不对，可以用默认值，或者报错。这里通常忽略错误使用默认值即可
	}

	// 2. 调用 Service
	users, total, err := service.ListUsers(c.Request.Context(), p)
	if err != nil {
		response.Fail(c, err)
		return
	}

	// 3. 组装响应
	// 步骤 A: 把 model.User 转成 serializer.User
	serializedUsers := serializer.BuildUsers(users)

	// 步骤 B: 把 list 和 total 包装成 DataList
	data := serializer.BuildDataList(serializedUsers, total, p.GetPage(), p.GetSize())

	response.Success(c, data)
}

// Show 获取当前用户详情
// @Summary 获取当前用户详情
// @Description 获取当前登录用户的详细信息
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} response.Response{data=serializer.UserResp} "成功返回"
// @Failure 401 {object} response.Response "未授权"
// @Router /user/profile [get]
func (u *UserController) Show(c *gin.Context) {

	// 1. 获取上下文中的 UserID (注意类型断言)
	// 假设中间件里存的是 uint 类型
	//value, exists := c.Get("userID")
	//if !exists {
	//	response.Fail(c, util.NewBizErr("Unauthorized", nil))
	//	return
	//}
	//
	//userID, ok := value.(int64)
	//if !ok {
	//	response.Fail(c, util.NewBizErr("Token 解析异常", nil))
	//	return
	//}

	userID := util.GetUserID(c)
	user, err := service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, serializer.BuildUser(*user))
}

// Created 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags User
// @Accept  json
// @Produce  json
// @Param request body request.RegisterReq true "注册参数"
// @Success 200 {object} response.Response{data=serializer.UserResp} "注册成功"
// @Failure 400 {object} response.Response "参数错误"
// @Router /register [post]
func (u *UserController) Created(c *gin.Context) {
	// 1. 参数绑定
	var req request.RegisterReq
	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, err)
		return
	}
	// 2. 调用 Service
	user, err := service.RegisterService(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, serializer.BuildUser(*user))
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取 Token
// @Tags User
// @Accept  json
// @Produce  json
// @Param request body request.LoginReq true "登录参数"
// @Success 200 {object} response.Response{data=map[string]string} "登录成功"
// @Failure 400 {object} response.Response "账号或密码错误"
// @Router /login [post]
func (u *UserController) Login(c *gin.Context) {

	var req request.LoginReq

	if err := c.ShouldBind(&req); err != nil {
		response.Fail(c, err)
		return
	}

	token, err := service.LoginService(c.Request.Context(), req.Account, req.Password)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, map[string]string{"token": token})
}
