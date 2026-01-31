package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"golang.org/x/crypto/bcrypt"
	"strings"
	"test/internal/dao"
	"test/internal/model"
	"test/pkg/config"
	"test/pkg/database"
	"test/pkg/jwt"
	myredis "test/pkg/redis"
	"test/pkg/util"
)

func GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	// ===========================
	// 1️⃣ 第一步：查询 Redis 缓存
	// ===========================
	// 定义一个唯一的 Key，例如 "user:info:1001"
	cacheKey := fmt.Sprintf("user:info:%d", id)

	// 从 Redis 获取字符串
	val, err := myredis.RedisClient.Get(ctx, cacheKey).Result()

	// 如果 err == nil，说明缓存里有数据 (缓存命中)
	if err == nil {
		var user model.User
		// 反序列化：把 JSON 字符串变回 Struct
		if jsonErr := json.Unmarshal([]byte(val), &user); jsonErr == nil {
			// ✅ 直接返回缓存数据，不走数据库
			return &user, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		// 如果是 redis.Nil 说明单纯没查到，如果是其他错（比如 Redis 挂了），记录个日志
		// log.Error("Redis error", err)
		// 注意：Redis 挂了不能影响主业务，所以这里我们不 return error，而是继续往下走查数据库
	}

	// ===========================
	// 2️⃣ 第二步：查询数据库 (缓存未命中)
	// ===========================
	userDao := dao.NewUserDao(database.DB)
	user, err := userDao.GetUserByID(ctx, id)
	if err != nil {
		// 数据库也没查到，那就是真没了
		return nil, util.NewBizErr("UserNotFound", nil)
	}

	// ===========================
	// 3️⃣ 第三步：回写 Redis (缓存预热)
	// ===========================
	// 序列化：把 Struct 变成 JSON 字符串
	data, _ := json.Marshal(user)

	// 写入 Redis，设置过期时间 (比如 1 小时)
	// ⚠️ 面试考点：一定要设置过期时间，防止死数据占满内存
	myredis.RedisClient.Set(ctx, cacheKey, data, time.Hour*1).Err()

	return user, nil
}

func ListUsers(ctx context.Context, req util.PaginationReq) ([]model.User, int64, error) {
	userDao := dao.NewUserDao(database.DB)
	users, total, err := userDao.ListUsers(ctx, req.GetOffset(), req.GetSize())
	if err != nil {
		return nil, 0, util.NewBizErr("SystemBusy", nil)
	}
	return users, total, nil
}

func LoginService(ctx context.Context, account, password string) (string, error) {

	var user model.User
	userDao := dao.NewUserDao(database.DB)
	if _, err := userDao.GetUserByAccount(ctx, account); err != nil {
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

func RegisterService(ctx context.Context, account, password string) (*model.User, error) {

	// 1. 初始化 DAO (使用全局 DB)
	userDao := dao.NewUserDao(database.DB)

	// 2. 业务逻辑：检查账号是否存在
	exist, err := userDao.ExistOrNotByAccount(ctx, account)
	if err != nil {
		return nil, util.NewBizErr("SystemBusy", nil)
	}
	if exist {
		return nil, util.NewBizErr("UserAlreadyExists", map[string]interface{}{
			"Name": account,
		})
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.NewBizErr("PasswordHashedErr", nil)
	}

	user := model.User{
		Password: string(hashedPwd),
		Account:  account,
	}

	// 5. 调用 DAO 保存
	if err := userDao.CreateUser(ctx, &user); err != nil {
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
