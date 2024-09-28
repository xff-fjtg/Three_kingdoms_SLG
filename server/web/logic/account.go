package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/models"
	"Three_kingdoms_SLG/server/web/model"
	"Three_kingdoms_SLG/utils"
	"log"
	"time"
)

var DefaultAccountLogic = &AccountLogic{}

type AccountLogic struct {
}

func (l AccountLogic) Register(rq *model.RegisterReq) error {
	username := rq.Username
	user := &models.User{}
	result := global.DB.Where("username = ?", username).First(&user)
	count := result.RowsAffected // 查询到的记录数量
	if count != 0 {
		//有数据用户已经存在
		return common.New(utils.UserExist, "user is exist")
	} else {
		user.Mtime = time.Now()
		user.Ctime = time.Now()
		user.Username = rq.Username
		user.Passcode = utils.RandSeq(6)
		user.Passwd = utils.Password(rq.Password, user.Passcode)

		user.Hardware = rq.Hardware
		err := global.DB.Save(&user).Error
		if err != nil {
			log.Println("register fail:", err)
			return common.New(utils.DBError, "database wrong")
		}
	}
	log.Println("register success")
	return nil
}
