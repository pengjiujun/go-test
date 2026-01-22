package serializer

import (
	"test/internal/model"
	"test/pkg/util"
)

// ---------------------------------------------
// 👇 新增这个结构体，专门用于 Swagger 文档生成
// ---------------------------------------------

// UserDataList 专门用于 API 文档的包装结构
// 解决了 Swagger 无法解析 DataList{items=[]UserResp} 的问题
type UserDataList struct {
	Items []UserResp `json:"items"` // 明确指定这里是 UserResp 数组
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
}

type UserResp struct {
	ID        uint           `json:"id"`
	Account   string         `json:"account"`
	CreatedAt util.LocalTime `json:"created_at"`
	UpdatedAt util.LocalTime `json:"updated_at"`
}

func BuildUser(item model.User) UserResp {
	return UserResp{
		ID:        item.ID,
		Account:   item.Account,
		CreatedAt: util.LocalTime(item.CreatedAt),
		UpdatedAt: util.LocalTime(item.UpdatedAt),
	}
}

func BuildUsers(items []model.User) []UserResp {
	var users []UserResp
	for _, item := range items {
		users = append(users, BuildUser(item))
	}
	return users
}
