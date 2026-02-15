package model

import "gorm.io/gorm"

// LmDtsGame undefined
type LmDtsGame struct {
	gorm.Model
	State             int8    `json:"state" gorm:"state"`                             // 游戏状态：1:等待加入/进行中，2:倒计时开始（封盘），3:已结束（结算完成）
	KillerRoom        int64   `json:"killer_room" gorm:"killer_room"`                 // 杀手房间：本局被选中的“死亡房间”编号
	PreKillerRoom     int64   `json:"pre_killer_room" gorm:"pre_killer_room"`         // 上局杀手房间：前一局的死亡房间编号
	TotalPeople       int64   `json:"total_people" gorm:"total_people"`               // 总人数：参与本局游戏的总玩家数
	TotalAmount       float64 `json:"total_amount" gorm:"total_amount"`               // 总下注额：本局所有玩家下注的总金额
	TotalBonus        float64 `json:"total_bonus" gorm:"total_bonus"`                 // 总奖金：本局派发出的总奖励
	TotalKillerAmount float64 `json:"total_killer_amount" gorm:"total_killer_amount"` // 杀手位总额：在死亡房间内的玩家下注总额
	StartTime         int64   `json:"start_time" gorm:"start_time"`                   // 开始时间：Unix 时间戳
	EndTime           int64   `json:"end_time" gorm:"end_time"`                       // 结束时间：倒计时结束的时间戳
	// 在 Record 表里找 GameId，它引用了我表里的 ID
	Records []LmDtsRecord `gorm:"foreignKey:GameId;references:ID"`
}

// TableName 表名称
func (*LmDtsGame) TableName() string {
	return "lm_dts_game"
}
