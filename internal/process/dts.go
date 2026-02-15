package process

import (
	"context"
	"test/pkg/util"
	"time"
)

func Handle(ctx context.Context) {

	InitGame()

	// 2. 启动结算/状态机协程 (独立运行)
	util.GoSafe(func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				CalcHandle()
			case <-ctx.Done():
				return
			}
		}
	})

	// 2. 启动推送任务 (使用 GoSafe)
	util.GoSafe(func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				StartPushTask()
			case <-ctx.Done():
				return
			}
		}
	})

}
