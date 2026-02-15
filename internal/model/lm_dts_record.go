package model

import "gorm.io/gorm"

// LmDtsRecord undefined
type LmDtsRecord struct {
	gorm.Model
	GameId int64     `json:"game_id" gorm:"game_id"`
	Game   LmDtsGame `json:"game" gorm:"foreignKey:GameId;references:ID"`

	UserId      int64   `json:"user_id" gorm:"user_id"`
	RoomId      int64   `json:"room_id" gorm:"room_id"`           // 房间 ID：玩家选择进入的房间（1-5，对应金木水火土）
	Amount      float64 `json:"amount" gorm:"amount"`             // 下注金额
	PaymentType string  `json:"payment_type" gorm:"payment_type"` // 支付方式：例如余额、等
	State       int8    `json:"state" gorm:"state"`               // 状态：0:等待 1:胜 2:负 结算状态：0:等待中，1:胜利（未被杀），2:失败（被杀）
	KillerRoom  int64   `json:"killer_room" gorm:"killer_room"`   // 结算时的杀手房间
	Bonus       float64 `json:"bonus" gorm:"bonus"`               // 获得奖金
	Num         int8    `json:"num" gorm:"num"`                   //倍数/编号
}

// TableName 表名称
func (*LmDtsRecord) TableName() string {
	return "lm_dts_record"
}
