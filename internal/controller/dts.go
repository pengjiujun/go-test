package controller

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"test/internal/model"
	"test/internal/request"
	"test/internal/service"
	"test/internal/websocket"
	"test/pkg/database"
	"test/pkg/response"
	"test/pkg/util"
)

type DtsController struct{}

func NewDtsController() *DtsController {
	return &DtsController{}
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (dts DtsController) Init(c *gin.Context) {
	// 1. 获取 ID（调用封装好的工具）
	userID := util.GetUserID(c)
	var user model.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		response.Fail(c, util.NewBizErr("用户获取错误", nil))
		return
	}

	// 2. 获取当前游戏
	var dtsGame model.LmDtsGame
	if err := database.DB.Order("id desc").First(&dtsGame).Error; err != nil {
		response.Fail(c, util.NewBizErr("当前没有正在进行的游戏", nil))
		return
	}

	req := &service.JoinGameReq{
		GameID:   dtsGame.ID,
		UserID:   userID,
		RoomID:   0,   // 默认进 1 号房
		Amount:   0.0, // 默认下注 0
		Nickname: "New Player",
	}

	// 2. 调用逻辑，直接拿回 Redis 的数据
	_, err := service.JoinUserList(c.Request.Context(), *req)
	if err != nil {
		response.Fail(c, util.NewBizErr("操作失败", nil))
		return
	}

	// 3. 直接返回组合数据
	response.Success(c, gin.H{
		"balance": user.Amount,
		"game_id": dtsGame.ID,
		"user_id": userID,
	})

}

func (dts DtsController) Quit(c *gin.Context) {

	userID := util.GetUserID(c)

	gameIdStr := c.DefaultQuery("game_id", "0")

	// 参数说明：字符串, 进制(10进制), 位数(64位)
	gameId, err := strconv.ParseInt(gameIdStr, 10, 64)

	if err != nil {
		response.Fail(c, util.NewBizErr("无效的游戏ID格式", nil))
		return
	}

	var game model.LmDtsGame
	if err := database.DB.Where("id = ?", gameId).First(&game).Error; err != nil {
		response.Fail(c, util.NewBizErr("当前游戏不存在", nil))
		return
	}

	err = service.RemoveUserList(c.Request.Context(), gameId, userID)
	if err != nil {
		response.Fail(c, util.NewBizErr("操作失败", nil))
		return
	}

	response.Success(c, gin.H{})
}

func (dts DtsController) Join(c *gin.Context) {

	userID := util.GetUserID(c)

	var joinReq request.JoinReq
	if err := c.ShouldBind(&joinReq); err != nil {
		response.Fail(c, err)
		return
	}

	if service.HasLock(c.Request.Context()) {
		response.Fail(c, util.NewBizErr("结算中", nil))
		return
	}

	// 1. 事务外：快速初筛（使用普通的 DB，不加锁）
	var game model.LmDtsGame
	if err := database.DB.First(&game, joinReq.GameID).Error; err != nil {
		response.Fail(c, util.NewBizErr("游戏不存在", nil))
		return
	}

	if game.State == 3 {
		response.Fail(c, util.NewBizErr("游戏已结束", nil))
		return
	}

	// 2. 开启事务
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 3. 事务内：悲观锁读取（核心屏障）
		var user model.User

		// 只有这行代码能保证并发安全！
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&user, userID).Error; err != nil {
			return err
		}

		// 二次检查余额（以锁定的数据为准）
		if user.Amount < joinReq.Amount {
			return errors.New("金额不足")
		}

		// 4. 处理下注记录 (Upsert 逻辑)
		var record model.LmDtsRecord
		result := tx.Where("user_id = ? AND game_id = ?", userID, game.ID).First(&record)

		var newTotalAmount float64
		if result.Error == nil {
			// 已有记录：累加金额并更新房间
			newTotalAmount = record.Amount + joinReq.Amount
			if err := tx.Model(&record).Updates(map[string]interface{}{
				"room_id": joinReq.RoomID,
				"amount":  newTotalAmount,
			}).Error; err != nil {
				return err
			}
		} else if errors.Is(result.Error, gorm.ErrRecordNotFound) {

			newTotalAmount = joinReq.Amount
			// 无记录：创建新记录
			if joinReq.Amount <= 0 {
				return errors.New("金额错误")
			}
			newRecord := model.LmDtsRecord{
				GameId: int64(joinReq.GameID),
				UserId: userID,
				RoomId: int64(joinReq.RoomID),
				Amount: joinReq.Amount,
			}
			if err := tx.Create(&newRecord).Error; err != nil {
				return err
			}
		} else {
			return result.Error
		}

		// 5. 扣除用户余额
		if err := tx.Model(&user).Update("amount", gorm.Expr("amount - ?", joinReq.Amount)).Error; err != nil {
			return err
		}

		err := service.UpdateGame(tx, &game)
		if err != nil {
			return err
		}

		req := &service.JoinGameReq{
			GameID:   game.ID,
			UserID:   userID,
			RoomID:   joinReq.RoomID, // 默认进 1 号房
			Amount:   newTotalAmount, // 默认下注 0
			Nickname: "New Player",
		}

		err = service.AddUserList(c.Request.Context(), req)
		if err != nil {
			return err
		}

		return nil

	})

	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{})

}

func (dts DtsController) Ws(c *gin.Context) {

	uid := util.GetUserID(c)
	fmt.Println(uid)
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &websocket.Client{
		ID:   uid,
		Send: make(chan []byte, 256),
	}

	websocket.GlobalHub.Register(uid, client)

	defer func() {
		websocket.GlobalHub.Unregister(uid) // 对应 onClose: 删除
		conn.Close()
	}()

	for msg := range client.Send {
		conn.WriteMessage(ws.TextMessage, msg)
	}

}
