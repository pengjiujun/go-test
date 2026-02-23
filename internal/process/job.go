package process

import (
	"context"
	"encoding/json"

	"github.com/shopspring/decimal"
	"test/pkg/redis"
)

func PushBonusJob(recordID uint, userID int64, totalAmount decimal.Decimal) {
	jobData := map[string]interface{}{
		"record_id": recordID,
		"user_id":   userID,
		"amount":    totalAmount.InexactFloat64(),
	}
	payload, _ := json.Marshal(jobData)
	// 推送到 Redis 队列
	redis.RedisClient.LPush(context.Background(), "game_bonus_queue", payload)
}
