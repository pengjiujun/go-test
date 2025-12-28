package service

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"test/internal/model"
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/jwt"
)

func LoginService(account, password string) (string, error) {

	var user model.User
	if err := database.DB.Where("account = ?", account).First(&user).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("密码错误")
	}

	token, err := jwt.NewJWT(config.Conf.Jwt.Secret, config.Conf.Jwt.Issuer, config.Conf.Jwt.ExpireSeconds).CreateToken(int64(user.ID))
	if err != nil {
		return "", errors.New("登录错误")
	}

	return token, nil
}
