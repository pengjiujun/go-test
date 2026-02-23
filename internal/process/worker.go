package process

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"test/internal/model"
	"test/pkg/database"
	"test/pkg/redis"
)

func StartBonusWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 使用 BRPOP 阻塞式获取任务，减少 CPU 消耗
			result, err := redis.RedisClient.BRPop(ctx, 0, "game_bonus_queue").Result()
			if err != nil {
				continue
			}

			var job map[string]interface{}
			json.Unmarshal([]byte(result[1]), &job)

			// 执行真正的加钱操作
			err = database.DB.Model(&model.User{}).
				Where("id = ?", job["user_id"]).
				UpdateColumn("amount", gorm.Expr("amount + ?", job["amount"])).Error

			if err != nil {
				// 失败处理：可以重新入队或记录错误日志记录
				fmt.Printf("发奖失败: %v", err)
			}
		}
	}
}
