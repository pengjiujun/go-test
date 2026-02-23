package process

import (
	"context"
	"encoding/json"
	"math"
	"test/internal/service"
	"test/internal/websocket"
	"time"
)

func StartPushTask() {

	// 1. 获取最新游戏 ID
	gameID, _ := service.GetLastGameId(context.Background())
	if gameID == 0 {
		return
	}
	// 2. 获取游戏基础数据
	game, _ := service.GetGame(gameID)
	userList, _ := service.GetUserList(context.Background(), int64(game.ID))
	// 对应 refreshData，准备公共部分

	// 3. 广播给所有在线用户
	clients := websocket.GlobalHub.GetAllClients()
	for _, client := range clients {
		// 获取该用户的个性化数据
		userData, err := service.GetUserGameData(context.Background(), int64(game.ID), client.ID)
		if err != nil {
			continue
		}
		pushData := map[string]interface{}{
			"dts_data": map[string]interface{}{
				"user_id":     client.ID,
				"user_bonus":  userData.Bonus,
				"user_amount": userData.Amount,

				"game_type":           1,
				"game_id":             game.ID,
				"start_time":          game.StartTime, //开始时间
				"end_time":            game.EndTime,   //结束时间
				"state":               game.State,     // 状态：1:进行中 2:开始倒计时 3:结束
				"timer":               math.Max(0, float64(game.EndTime-time.Now().Unix())),
				"killer_room":         game.KillerRoom,
				"pre_killer_room":     game.PreKillerRoom,
				"join_people":         len(userList), // 加入的人
				"max_people":          3,             //房间人数限制
				"total_killer_amount": game.TotalKillerAmount,
				"user_list":           userList,
				"room_list":           service.CalcRoomAmount(userList),
				"timestamp":           time.Now().Unix(),
			},
		}

		payload, _ := json.Marshal(pushData)

		// 异步发送，不阻塞当前广播循环
		select {
		case client.Send <- payload:
		default:
			// 通道满说明网络卡，可跳过
		}
	}

}
