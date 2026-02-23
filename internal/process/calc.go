package process

import (
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"math/rand"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/database"
	"test/pkg/redis"
	"time"
)

// 全局随机源，初始化一次
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func getLastGame() (*model.LmDtsGame, error) {
	//var DtsGame model.LmDtsGame
	//err := database.DB.Preload("Records").Where("state = ?", 2).Where("end_time <= ?", time.Now()).First(&DtsGame).Error
	//return &DtsGame, err

	var games []model.LmDtsGame // 定义一个切片

	// 使用 Find 代替 First，找不到记录时 err 为 nil，且不会打印日志
	err := database.DB.Preload("Records").
		Where("state = ?", 2).
		Where("end_time <= ?", time.Now().Unix()).
		Limit(1).
		Find(&games).Error

	if err != nil {
		return nil, err // 真正的数据库连接错误等
	}

	if len(games) == 0 {
		return nil, errors.New("未找到记录") // 没找到记录，返回空指针且不报错
	}

	return &games[0], nil
}

func CalcHandle() {

	// 1. 增加分布式锁，防止 Ticker 导致重叠结算
	lockKey := "game_dts_calc_lock"
	ok, err := redis.RedisClient.SetNX(context.Background(), lockKey, "1", 10*time.Second).Result()
	if err != nil || !ok {
		return
	}
	defer redis.RedisClient.Del(context.Background(), lockKey)

	game, err := getLastGame()

	if err != nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	killerRoom := calc(game)
	//等待前端的动画
	time.Sleep(time.Second)
	//添加新的一期
	addGame(killerRoom)
	// 删除上期缓存数据
	err = service.DeleteUserList(context.Background(), int64(game.ID))
	if err != nil {
		return
	}
}

func calc(game *model.LmDtsGame) int64 {

	killRoom := getKillerRoom(game)

	//被刀房间的所有投注
	//所有房间的投注
	var totalKillerAmount, totalAmount float64

	//query := database.DB.Model(model.LmDtsRecord{}).Where("game_id = ?", game.ID)
	//query.Session(&gorm.Session{}).Where("room_id IN ?", killRoom).Pluck("SUM(amount)", &totalKillerAmount)
	//query.Where("room_id IN ?", killRoom).Pluck("SUM(amount)", &totalKillerAmount)
	//query.Pluck("SUM(amount)", &totalAmount)

	// 性能优化：用一条查询获取两个统计值
	err := database.DB.Model(&model.LmDtsRecord{}).Where("game_id = ?", game.ID).
		Select("SUM(CASE WHEN room_id = ? THEN amount ELSE 0 END) as killer_amount, SUM(amount) as total_amount", killRoom).
		Row().Scan(&totalKillerAmount, &totalAmount)
	if err != nil {
		return 0
	}

	// 转换为 Decimal 进行后续运算
	dTotalAmount := decimal.NewFromFloat(totalAmount)
	dKillerAmount := decimal.NewFromFloat(totalKillerAmount)
	dDivisor := dTotalAmount.Sub(dKillerAmount) // 胜出者总投注额

	err = database.DB.Transaction(func(tx *gorm.DB) error {
		var totalPeople int64
		var totalBonus decimal.Decimal = decimal.NewFromInt(0)

		//// 这里的 divisor 要防止为 0
		//divisor := totalAmount - totalKillerAmount

		for _, record := range game.Records {

			totalPeople++
			record.KillerRoom = killRoom

			if record.RoomId == killRoom {
				// 判定为失败（被杀）
				record.State = 2
				//database.DB.Save(&record)
				// ✅ 必须使用 tx!
				if err := tx.Save(&record).Error; err != nil {
					return err
				}
				continue
			}

			// 计算奖金 (防止除以0)
			bonus := decimal.Zero
			if dDivisor.GreaterThan(decimal.Zero) {
				personalAmt := decimal.NewFromFloat(record.Amount)
				// bonus = killerAmount * 0.9 * (personalAmt / divisor)
				bonus = dKillerAmount.Mul(decimal.NewFromFloat(0.9)).Mul(personalAmt.Div(dDivisor))
			}
			//累计共产生多少奖金
			totalBonus = totalBonus.Add(bonus)

			//var user model.User
			//if err := tx.First(&user, record.UserId).Error; err != nil {
			//	return err
			//}
			//user.Amount = bonus + record.Amount
			//database.DB.Save(&user)

			// ✅ 原子操作：增加余额（不要先 First 再 Save，高并发下会覆盖）
			// ✅ 原子操作：退回本金 + 增加奖金，防止并发覆盖
			totalReturn := bonus.Add(decimal.NewFromFloat(record.Amount))
			if err = tx.Model(&model.User{}).Where("id = ?", record.UserId).
				UpdateColumn("amount", gorm.Expr("amount + ?", totalReturn.InexactFloat64())).Error; err != nil {
				return err
			}
			record.Bonus = bonus.InexactFloat64()
			record.State = 1
			//database.DB.Save(&record)

			if err = tx.Save(&record).Error; err != nil {
				return err
			}
		}

		//Save所有字段更新 典型场景 完整对象保存
		// 3. 完整更新游戏主表参数
		// 使用 Updates 确保所有统计字段一次性写入
		// 使用结构体（零值字段不会被更新）
		// db.Model(&User{ID: 1}).Updates(User{Name: "王五", Age: 0})
		// 生成的 SQL: UPDATE users SET name='王五' WHERE id=1; (age=0 被忽略)
		//需要更新零值：使用 map 形式的 Updates 或 Update
		// 将年龄设为 0
		//db.Model(&user).Updates(map[string]interface{}{"age": 0})
		return tx.Model(game).Updates(map[string]interface{}{
			"state":               3,                           // 3:已结束（结算完成）
			"killer_room":         killRoom,                    // 本局杀手房间
			"total_amount":        totalAmount,                 // 总下注额
			"total_people":        totalPeople,                 // 总参与人数
			"total_bonus":         totalBonus.InexactFloat64(), // 本局总派发奖金
			"total_killer_amount": totalKillerAmount,           // 杀手位总额
			"end_time":            time.Now().Unix(),           // 记录实际结束时间
		}).Error
	})

	if err != nil {
		return 0
	}

	return killRoom

}

func InitGame() {
	var count int64
	if err := database.DB.Model(model.LmDtsGame{}).Count(&count).Error; err != nil {
		panic(err)
	}

	if count == 0 {
		addGame(0)
	}
}

func addGame(preKillerRoom int64) {

	dtsGame := model.LmDtsGame{
		State:         1,
		KillerRoom:    0,
		PreKillerRoom: preKillerRoom, //上局杀手房间
		TotalPeople:   0,
		TotalAmount:   0,
		TotalBonus:    0,
	}

	if err := database.DB.Create(&dtsGame).Error; err != nil {
		panic(err)
	}

	service.SetLastGameId(context.Background(), dtsGame.ID)
}

func getKillerRoom(game *model.LmDtsGame) int64 {

	var roomIds []int64

	database.DB.Model(&model.LmDtsRecord{}).
		Where("game_id = ?", game.ID).
		Distinct().
		Pluck("room_id", &roomIds)
	if len(roomIds) == 0 {
		return 0
	}

	index := r.Intn(len(roomIds))
	return roomIds[index]
}
