package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
	"test/internal/model"
	"test/pkg/database"
	myredis "test/pkg/redis"
	"time"
)

type DtsUserCache struct {
	UserID   int64   `json:"user_id"`
	Nickname string  `json:"nickname"`
	GameID   uint    `json:"game_id"`
	RoomID   int     `json:"room_id"`
	Amount   float64 `json:"amount"`
	Bonus    float64 `json:"bonus"`
}

// 1. 定义一个专门的请求结构体
type JoinGameReq struct {
	GameID   uint
	UserID   int64
	RoomID   int
	Amount   float64
	Nickname string
}

const Key = "dts_game_user_list"

func JoinUserList(ctx context.Context, req JoinGameReq) (*DtsUserCache, error) {
	key := fmt.Sprintf("%s:%d", Key, req.GameID)
	field := strconv.FormatInt(req.UserID, 10)

	// 1. 尝试从 Redis 获取已有数据
	res, err := myredis.RedisClient.HGet(ctx, key, field).Result()
	if err == nil {
		// 如果已经存在，直接解析并返回，不要重复 HSet
		var cache DtsUserCache
		if err = json.Unmarshal([]byte(res), &cache); err != nil {
			return nil, fmt.Errorf("json unmarshal error: %w", err) // 使用 %w 包装错误
		}
		return &cache, nil
	}
	if !errors.Is(err, redis.Nil) {
		// 如果是 Redis 出错（非数据不存在），则向上抛出错误
		return nil, err
	}

	// 2. 如果不存在，构造新数据
	cache := &DtsUserCache{
		UserID:   req.UserID,
		Nickname: req.Nickname,
		GameID:   req.GameID,
		RoomID:   req.RoomID, // 初始房间
		Amount:   req.Amount, // 初始金额
	}

	// 3. 写入 Redis
	dataBytes, _ := json.Marshal(cache)

	if err = myredis.RedisClient.HSet(ctx, key, field, dataBytes).Err(); err != nil {
		return nil, err
	}

	return cache, nil
}

func RemoveUserList(ctx context.Context, GameID, UserID int64) error {
	key := fmt.Sprintf("%s:%d", Key, GameID)
	field := strconv.FormatInt(UserID, 10)
	_, err := myredis.RedisClient.HDel(ctx, key, field).Result()
	return err
}

func DeleteUserList(ctx context.Context, GameID int64) error {
	key := fmt.Sprintf("%s:%d", Key, GameID)
	_, err := myredis.RedisClient.HDel(ctx, key).Result()
	return err
}

func AddUserList(ctx context.Context, req *JoinGameReq) error {
	key := fmt.Sprintf("%s:%d", Key, req.GameID)
	field := strconv.FormatInt(req.UserID, 10)
	// 2. 如果不存在，构造新数据
	cache := &DtsUserCache{
		UserID:   req.UserID,
		Nickname: req.Nickname,
		GameID:   req.GameID,
		RoomID:   req.RoomID,
		Amount:   req.Amount,
	}
	// 3. 写入 Redis
	dataBytes, _ := json.Marshal(cache)
	return myredis.RedisClient.HSet(ctx, key, field, dataBytes).Err()
}

func GetUserList(ctx context.Context, GameID int64) ([]DtsUserCache, error) {
	key := fmt.Sprintf("%s:%d", Key, GameID)
	data, err := myredis.RedisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	// 3. 预分配结果切片（容量设为 data 的长度，减少扩容开销）
	userList := make([]DtsUserCache, 0, len(data))

	// 4. 遍历数据 (对应 PHP 的 array_map + array_filter)
	for _, valStr := range data {
		var u DtsUserCache
		// 反序列化 JSON (对应 json_decode)
		if err := json.Unmarshal([]byte(valStr), &u); err != nil {
			continue // 解析失败跳过
		}

		// 5. 过滤逻辑 (对应 PHP 的 array_filter)
		// 只有 User 存在且 RoomID 不为 0 时才加入结果集
		if u.RoomID > 0 {
			userList = append(userList, u)
		}
	}

	return userList, nil
}

func HasLock(ctx context.Context) bool {
	exists, _ := myredis.RedisClient.Exists(ctx, "game_dts_lock_flag").Result()
	if exists == 0 {
		return false
	}
	return true
}

const (
	MaxPeople = 2
	Duration  = 30
)

func UpdateGame(tx *gorm.DB, game *model.LmDtsGame) error {
	// 1. 状态校验：只有进行中(1)的场次才能触发倒计时
	if game.State != 1 {
		return nil
	}

	// 2. 统计参与人数 (对应 $game->records()->count())
	var total int64
	// 使用关联统计，不需要把记录全查出来
	err := tx.Model(&model.LmDtsRecord{}).Where("game_id = ?", game.ID).Count(&total).Error
	if err != nil {
		return err
	}

	// 3. 检查是否达到人数阈值
	if total >= int64(MaxPeople) {
		now := time.Now().Unix()

		// 4. 更新状态为倒计时中(2)
		// 使用 Map 更新以确保所有字段（包括可能为0的字段）都能被正确写入
		err := tx.Model(game).Updates(map[string]interface{}{
			"state":      2,
			"start_time": now,
			"end_time":   now + int64(Duration),
		}).Error

		if err != nil {
			return err
		}
	}

	return nil
}

func SetLastGameId(ctx context.Context, ID uint) {
	myredis.RedisClient.Set(ctx, "game_dts_last_game_id", ID, 0)
}

func GetLastGameId(ctx context.Context) (uint, error) {
	result, err := myredis.RedisClient.Get(ctx, "game_dts_last_game_id").Result()
	if err != nil {
		return 0, err
	}
	resultInt, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(resultInt), nil
}

func GetGame(gameID uint) (*model.LmDtsGame, error) {
	var game model.LmDtsGame
	if err := database.DB.Where("id = ?", gameID).First(&game).Error; err != nil {
		return nil, err
	}
	return &game, nil
}

type RoomAmount struct {
	RoomID int             `json:"room_id"`
	Amount decimal.Decimal `json:"amount"` // 使用 decimal 类型保证精度
}

func CalcRoomAmount(userList []DtsUserCache) []RoomAmount {

	// 1. 初始化一个 Map 用于存放每个房间的金额累加
	// key 是房间 ID，value 是累加的金额
	roomMap := make(map[int]decimal.Decimal)

	// 2. 只需要遍历一次用户列表 (O(N) 时间复杂度)
	for _, item := range userList {
		if item.RoomID > 0 {
			// 获取当前房间已有的金额（如果不存在则是 0 值）
			currentTotal := roomMap[int(item.RoomID)]
			// ❌ 错误写法：currentTotal.Add(item.Amount)
			// ✅ 正确写法：使用 NewFromFloat 将 float64 转为 Decimal
			changeAmount := decimal.NewFromFloat(item.Amount)

			// 执行加法并存回 Map
			roomMap[int(item.RoomID)] = currentTotal.Add(changeAmount)
		}
	}

	// 3. 构建返回数据 (固定 1-9 号房间)
	results := make([]RoomAmount, 0, 9)
	for i := 1; i <= 9; i++ {
		amount, exists := roomMap[i]
		if !exists {
			amount = decimal.NewFromInt(0) // 如果该房间没人，金额为 0
		}

		results = append(results, RoomAmount{
			RoomID: i,
			Amount: amount,
		})
	}

	return results

}

func GetUserGameData(ctx context.Context, GameID, UserID int64) (*DtsUserCache, error) {

	key := fmt.Sprintf("%s:%d", Key, GameID)
	field := strconv.FormatInt(UserID, 10)

	data, err := myredis.RedisClient.HGet(ctx, key, field).Result()
	if err != nil {
		return nil, err
	}
	var userList DtsUserCache
	if err := json.Unmarshal([]byte(data), &userList); err != nil {
		return nil, err
	}
	return &userList, nil
}
