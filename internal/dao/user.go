package dao

import (
	"context"
	"gorm.io/gorm"
	"test/internal/model"
)

// UserDao 定义 DAO 结构体
type UserDao struct {
	db *gorm.DB
}

// NewUserDao 构造函数 (依赖注入的雏形)
// 传入 *gorm.DB，方便以后换数据库连接或者 Mock 测试
func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{db: db}
}

// ExistOrNotByAccount 检查账号是否存在
func (dao *UserDao) ExistOrNotByAccount(ctx context.Context, account string) (bool, error) {
	var count int64
	// WithContext 是 Go 后端开发的良好习惯，用于链路追踪和超时控制
	err := dao.db.WithContext(ctx).Model(&model.User{}).Where("account = ?", account).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateUser 创建用户
func (dao *UserDao) CreateUser(ctx context.Context, user *model.User) error {
	return dao.db.WithContext(ctx).Create(user).Error
}

// GetUserByAccount 根据账号获取用户信息
func (dao *UserDao) GetUserByAccount(ctx context.Context, account string) (*model.User, error) {
	var user model.User
	err := dao.db.WithContext(ctx).Where("account = ?", account).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (dao *UserDao) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := dao.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (dao *UserDao) ListUsers(ctx context.Context, offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	db := dao.db.WithContext(ctx).Model(&model.User{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.Order("id desc").Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}
