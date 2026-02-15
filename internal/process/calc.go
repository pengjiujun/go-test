package process

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"math/rand"
	"test/internal/model"
	"test/internal/service"
	"test/pkg/database"
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
		Where("end_time <= ?", time.Now()).
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

	err = database.DB.Transaction(func(tx *gorm.DB) error {

		var totalPeople int64
		var totalBonus float64

		// 这里的 divisor 要防止为 0
		divisor := totalAmount - totalKillerAmount

		for _, record := range game.Records {

			totalPeople++
			record.KillerRoom = killRoom

			if record.RoomId == killRoom {
				record.State = 2
				//database.DB.Save(&record)
				// ✅ 必须使用 tx!
				if err := tx.Save(&record).Error; err != nil {
					return err
				}
				continue
			}

			// 计算奖金 (防止除以0)
			bonus := 0.0
			if divisor > 0 {
				bonus = totalKillerAmount * 0.9 * (record.Amount / divisor)
			}

			//累计共产生多少奖金
			totalBonus += bonus

			//var user model.User
			//if err := tx.First(&user, record.UserId).Error; err != nil {
			//	return err
			//}
			//user.Amount = bonus + record.Amount
			//database.DB.Save(&user)

			// ✅ 原子操作：增加余额（不要先 First 再 Save，高并发下会覆盖）
			if err = tx.Model(&model.User{}).Where("id = ?", record.UserId).
				UpdateColumn("amount", gorm.Expr("amount + ?", bonus+record.Amount)).Error; err != nil {
				return err
			}

			record.Bonus = bonus
			record.State = 1
			//database.DB.Save(&record)

			if err = tx.Save(&record).Error; err != nil {
				return err
			}
		}

		game.State = 3
		game.EndTime = time.Now().Unix()
		game.TotalAmount = totalAmount
		game.TotalPeople = totalPeople
		game.TotalBonus = totalBonus
		game.KillerRoom = killRoom
		game.TotalKillerAmount = totalKillerAmount
		//database.DB.Save(&game)
		return tx.Save(&game).Error

		// 返回 nil 提交事务
		//return nil
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
