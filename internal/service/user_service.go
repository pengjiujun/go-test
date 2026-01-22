package service

import (
	"golang.org/x/crypto/bcrypt"
	"strings"
	"test/internal/model"
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/jwt"
	"test/pkg/util"
	"unicode/utf8"
)

func GetUserByID(id int64) (*model.User, error) {
	var user model.User
	// 这里未来可以加 Redis 缓存！
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, util.NewBizErr("UserNotFound", nil)
	}
	return &user, nil
}

func ListUsers(req util.PaginationReq) ([]model.User, int64, error) {

	var users []model.User
	var total int64

	// 1. 准备查询构建器 (如果有搜索条件，放在这里，比如 db = db.Where("name LIKE ?", ...))
	db := database.DB.Model(&model.User{})

	// 2. 先查总数 (注意：要在 Limit 之前查)
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 3. 再查数据 (应用分页)
	// Scopes 是 GORM 的高级用法，也可以直接写 .Offset().Limit()
	// 正确：先 Order，再 Limit/Offset，最后 Find
	err := db.Order("id desc"). // 1. 先决定排序 (很重要！分页必须先排序)
					Offset(req.GetOffset()). // 2. 再偏移
					Limit(req.GetSize()).    // 3. 再截取
					Find(&users).            // 4. 最后执行查询！
					Error

	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func LoginService(account, password string) (string, error) {

	var user model.User
	if err := database.DB.Where("account = ?", account).First(&user).Error; err != nil {
		return "", util.NewBizErr("UserNotFound", nil)
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", util.NewBizErr("PasswordIncorrect", nil)
	}

	token, err := jwt.NewJWT(config.Conf.Jwt.Secret, config.Conf.Jwt.Issuer, config.Conf.Jwt.ExpireSeconds).CreateToken(int64(user.ID))
	if err != nil {
		return "", util.NewBizErr("TokenGenerateFailed", nil)
	}

	return token, nil
}

func RegisterService(account, password string) (*model.User, error) {
	// 1. 应用层先检查一次 (为了过滤掉大部分正常情况下的重复，减少 DB 写入压力)
	var count int64
	database.DB.Model(&model.User{}).Where("account = ?", account).Count(&count)
	if count > 0 {
		return nil, util.NewBizErr("UserAlreadyExists", map[string]interface{}{
			"Name": account,
		})
	}

	if utf8.RuneCount([]byte(password)) < 6 {
		return nil, util.NewBizErr("PasswordTooShort", nil)
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.NewBizErr("PasswordHashedErr", nil)
	}

	user := model.User{
		Password: string(hashedPwd),
		Account:  account,
	}

	if err := database.DB.Create(&user).Error; err != nil {

		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, util.NewBizErr("UserAlreadyExists", map[string]interface{}{
				"Name": account,
			})
		}

		// 其他错误才是系统繁忙
		return nil, util.NewBizErr("SystemBusy", nil)
	}

	return &user, nil
}
