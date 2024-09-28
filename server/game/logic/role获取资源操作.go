package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/gameConfig"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

var RoleService = &roleService{}

type roleService struct {
}

func (r *roleService) EnterServer(uid int, rsp *model.EnterServerRsp, conn net.WSConn) error {
	// 开始事务
	tx := global.DB.Begin()
	if tx.Error != nil {
		return common.New(utils.DBError, "开启事务失败")
	}
	defer func() {
		// 如果操作失败，回滚事务
		if r := recover(); r != nil {
			tx.Rollback()
			log.Println("事务回滚 due to panic:", r)
		}
	}()
	role := &data.Role{}
	err := tx.Where("uid = ?", uid).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.New(utils.RoleNotExist, "角色不存在")
		}
		log.Println("查询角色出错", err)
		return common.New(utils.DBError, "查询角色出错")
	}
	rid := role.RId
	roleRes := &data.RoleRes{}
	//查询资源
	result := tx.Where("rid = ?", rid).First(&roleRes)
	if result.RowsAffected == 0 {
		//资源不存在  加载初始资源
		roleRes = &data.RoleRes{
			RId:    role.RId,
			Wood:   gameConfig.Base.Role.Wood,
			Iron:   gameConfig.Base.Role.Iron,
			Stone:  gameConfig.Base.Role.Stone,
			Grain:  gameConfig.Base.Role.Grain,
			Gold:   gameConfig.Base.Role.Gold,
			Decree: gameConfig.Base.Role.Decree,
		}
		err := tx.Create(&roleRes).Error
		if err != nil {
			log.Println("插入角色资源出错", err)
			tx.Rollback()
			return common.New(utils.DBError, "插入角色资源出错")
		}
	} else if result.Error != nil {
		log.Println("查询角色资源出错", result.Error)
		tx.Rollback()
		return common.New(utils.DBError, "查询角色资源出错")
	}
	rsp.RoleRes = roleRes.ToModel().(model.RoleRes)
	// 现在 rspObj.RoleRes 就是一个 model.RoleRes 类型的实例，
	//只包含需要存储的字段
	rsp.Role = role.ToModel().(model.Role)
	rsp.Token, _ = utils.Award(rid)
	rsp.Time = time.Now().UnixNano() / 1e6
	conn.SetProperty("role", role)
	//初始化玩家属性
	if err := RoleAttrService.TryCreate(rid, conn, tx); err != nil {
		tx.Rollback()
		return common.New(utils.DBError, "初始化玩家属性fail")
	}
	//初始化玩家主城
	if err := RoleCity.InitCity(rid, role.NickName, conn, tx); err != nil {
		tx.Rollback()
		return common.New(utils.DBError, "初始化玩家city fail")
	}
	// 所有操作成功，提交事务
	if err := tx.Commit().Error; err != nil {
		log.Println("提交事务失败", err)
		return common.New(utils.DBError, "提交事务失败")
	}
	//进入游戏 放入连接
	net.Mgr.RoleEnter(conn, role.RId)
	return nil
}

func (r *roleService) GetRoleRes(rid int) (model.RoleRes, error) {
	fmt.Println("yes")
	roleRes := &data.RoleRes{}
	//查询资源
	count := global.DB.Where("rid = ?", rid).First(&roleRes).RowsAffected
	if count == 0 {
		log.Println("查询角色资源出错")
		return model.RoleRes{}, common.New(utils.DBError, "查询角色资源出错")
	}
	//count!=0 查到了
	return roleRes.ToModel().(model.RoleRes), nil
}

func (r *roleService) GetYield(rid int) data.Yield {
	//产量+建筑+城市固定收益 = 最终的产量

	rbYield := RoleBuild.GetYield(rid)
	rcYield := CityFacilityService.GetYield(rid)
	var y data.Yield

	y.Gold = rbYield.Gold + rcYield.Gold + gameConfig.Base.Role.GoldYield
	y.Stone = rbYield.Stone + rcYield.Stone + gameConfig.Base.Role.StoneYield
	y.Iron = rbYield.Iron + rcYield.Iron + gameConfig.Base.Role.IronYield
	y.Grain = rbYield.Grain + rcYield.Grain + gameConfig.Base.Role.GrainYield
	y.Wood = rbYield.Wood + rcYield.Wood + gameConfig.Base.Role.WoodYield

	return y
}

func (r *roleService) Get(rid int) *data.Role {
	role := &data.Role{}
	result := global.DB.Where("rid=?", rid).Find(&role)
	if result.Error != nil {
		log.Println("查询角色出错", result.Error)
		return nil
	}

	return role
}
