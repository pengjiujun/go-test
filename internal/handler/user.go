package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/database"
	"time"
)

type User struct{}

type reqUser struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
}

type respUser struct {
	Id        uint      `json:"id"`
	Account   string    `json:"account"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type respLogin struct {
	Token string   `json:"token"`
	User  respUser `json:"user"`
}

type reqUserLogin struct {
	Account  string `json:"account" form:"account"`
	Password string `json:"password" form:"password"`
}

func (u *User) Index(c *gin.Context) {
	var users []model.User
	if err := database.DB.Find(&users).Error; err != nil {
		zap.L().Info("查询用户错误")
		c.JSON(422, ErrorResponse(422, "查询用户错误"))
		return
	}
	c.JSON(200, SuccessResponse(users))
}

func (u *User) Show(c *gin.Context) {

	userID, _ := c.Get("userID")
	fmt.Println(userID)
}

func (u *User) Created(c *gin.Context) {

	var req reqUser
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(422, ErrorResponse(422, err.Error()))
		return
	}

	var count int64
	if err := database.DB.Model(&model.User{}).Where("account = ?", req.Account).Count(&count).Error; err != nil {
		c.JSON(422, ErrorResponse(422, "xx"+err.Error()))
		return
	}

	if count > 0 {
		c.JSON(422, ErrorResponse(422, "账号已存在"))
		return
	}

	// 注册时：加密密码
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		Password: string(hashedPwd),
		Account:  req.Account,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(422, ErrorResponse(422, "创建失败"))
		return
	}

	c.JSON(200, SuccessResponse(respUser{
		Id:        user.ID,
		Account:   user.Account,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}))
}

func (u *User) Login(c *gin.Context) {

	var req reqUserLogin
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(422, ErrorResponse(422, err.Error()))
		return
	}

	token, err := service.LoginService(req.Account, req.Password)
	if err != nil {
		c.JSON(422, ErrorResponse(422, err.Error()))
		return
	}

	c.JSON(200, SuccessResponse(gin.H{"token": token}))

}
