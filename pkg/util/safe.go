package util

import (
	"fmt"
	"runtime/debug"
)

// GoSafe 安全地启动一个协程
// 如果协程内发生 Panic，会捕获并打印堆栈，防止整个程序崩溃
func GoSafe(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// 1. 打印错误信息
				fmt.Printf("Recovered from panic: %v\n", r)
				// 2. 打印堆栈信息（非常重要，否则你不知道哪里错了）
				debug.PrintStack()
				// 3. 这里可以加报警通知（比如发送到钉钉/企业微信）
			}
		}()

		// 执行真正的业务逻辑
		fn()
	}()
}
