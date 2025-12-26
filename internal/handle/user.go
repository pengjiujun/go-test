package handle

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"test/internal/model"
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/jwt"
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
	user := model.User{
		Account:  req.Account,
		Password: req.Password,
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

	var user model.User
	if err := database.DB.Where("account = ?", req.Account).First(&user).Error; err != nil {
		c.JSON(422, ErrorResponse(422, "用户不存在"))
		return
	}

	if user.Password != req.Password {
		c.JSON(422, ErrorResponse(422, "密码错误"))
		return
	}

	newJwt := jwt.NewJWT(config.Conf.Jwt.Secret, config.Conf.Jwt.Issuer, config.Conf.Jwt.ExpireSeconds)
	token, err := newJwt.CreateToken(int64(user.ID))
	if err != nil {
		c.JSON(422, ErrorResponse(422, "登录错误"))
		return
	}

	c.JSON(200, SuccessResponse(respLogin{
		Token: token,
		User: respUser{
			Id:        user.ID,
			Account:   user.Account,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}))

}
