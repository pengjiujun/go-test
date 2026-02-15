package request

type JoinReq struct {
	GameID int     `json:"game_id" form:"game_id" binding:"required" label:"GameID"`
	RoomID int     `json:"room_id" form:"room_id" binding:"required" label:"RoomID"`
	Amount float64 `json:"amount" form:"amount" binding:"required" label:"Amount"`
}
