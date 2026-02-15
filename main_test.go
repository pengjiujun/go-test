// main_test.go
package main

import (
	"fmt"
	"test/internal/model"
	"test/pkg/config"
	"test/pkg/database"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"golang.org/x/crypto/bcrypt"
)

// TestSeedData 是数据填充的入口
// 运行命令: go test -v -run TestSeedData
func TestSeedData(t *testing.T) {
	// 1. 环境初始化
	config.Load()
	database.InitDb()

	fmt.Println("\n🚀 [Seed] 开始往数据库灌入模拟数据...")

	// 2. 执行填充
	seedUsers(t, 20)
	seedBanners(t, 10)

	fmt.Println("✅ [Seed] 数据填充大功告成！")
}

func seedUsers(t *testing.T, count int) {
	password, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	for i := 0; i < count; i++ {
		user := model.User{
			Account:  GenerateChinaPhone(), // 随机手机号
			Password: string(password),
		}
		if err := database.DB.Create(&user).Error; err != nil {
			t.Errorf("创建用户失败: %v", err)
		}
	}
	fmt.Printf("-> 已生成 %d 条用户数据\n", count)
}

func seedBanners(t *testing.T, count int) {
	for i := 0; i < count; i++ {
		banner := model.Banner{
			ImageUrl: gofakeit.ImageURL(800, 400),
			Sort:     gofakeit.Number(1, 100),
		}
		if err := database.DB.Create(&banner).Error; err != nil {
			t.Errorf("创建 Banner 失败: %v", err)
		}
	}
	fmt.Printf("-> 已生成 %d 条 Banner 数据\n", count)
}

func GenerateChinaPhone() string {
	// 常见的中国手机号开头
	prefixes := []string{"138", "139", "158", "188", "170", "199", "133"}
	prefix := prefixes[gofakeit.Number(0, len(prefixes)-1)]

	// 后面补齐 8 位数字
	return prefix + gofakeit.DigitN(8)
}
