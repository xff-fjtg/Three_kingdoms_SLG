package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"log"
	"sync"
)

// 开始先把所有联盟加载
var CoalitionService = &coalitionService{
	unions: make(map[int]*data.Coalition),
}

type coalitionService struct {
	mutex  sync.RWMutex
	unions map[int]*data.Coalition
}

// 么mber[1,2,3,4,5] json的字符串
func (c *coalitionService) Load() {
	rr := make([]*data.Coalition, 0)
	err := global.DB.Where("state = ?", data.UnionRunning).Find(&rr)
	if err.Error != nil {
		log.Println("coalitionService load error", err.Error)
	}
	for _, v := range rr {
		members := v.Members
		err := json.Unmarshal([]byte(members), &v.MemberArray)
		if err != nil {
			log.Println("coalitionService load error", err)
			return
		}
		//afterSer()
		c.unions[v.Id] = v
	}
}

func (c *coalitionService) List() ([]model.Union, error) {
	r := make([]model.Union, 0)
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for _, coalition := range c.unions {
		union := coalition.ToModel().(model.Union)
		//盟主和副盟主信息
		main := make([]model.Major, 0)
		if role := RoleService.Get(coalition.Chairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		if role := RoleService.Get(coalition.ViceChairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		union.Major = main
		r = append(r, union)
	}
	return r, nil
}

func (c *coalitionService) ListCoalition() []*data.Coalition {
	r := make([]*data.Coalition, 0)
	for _, coalition := range c.unions {
		r = append(r, coalition)
	}
	return r
}

func (c *coalitionService) Get(id int) (model.Union, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	coalition, ok := c.unions[id]
	if ok {
		union := coalition.ToModel().(model.Union)
		//盟主和副盟主信息
		main := make([]model.Major, 0)
		if role := RoleService.Get(coalition.Chairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		if role := RoleService.Get(coalition.ViceChairman); role != nil {
			m := model.Major{Name: role.NickName, RId: role.RId, Title: model.UnionChairman}
			main = append(main, m)
		}
		union.Major = main
		return union, nil
	}
	return model.Union{}, nil
}

func (c *coalitionService) GetCoalition(id int) *data.Coalition {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	coa, ok := c.unions[id]
	if ok {
		return coa
	}
	return nil
}

func (c *coalitionService) GetListApply(unionId int, state int) ([]model.ApplyItem, error) {
	applys := make([]data.CoalitionApply, 0)
	err := global.DB.
		Where("union_id = ? and state = ? ", unionId, state).
		Find(&applys)
	if err != nil {
		log.Println("coalitionService GetListApply find error", err)
		return nil, common.New(utils.DBError, "数据库错误")
	}
	ais := make([]model.ApplyItem, 0)
	for _, v := range applys {
		var ai model.ApplyItem
		ai.Id = v.ID
		role := RoleService.Get(v.RId)
		ai.NickName = role.NickName
		ai.RId = role.RId
		ais = append(ais, ai)
	}
	return ais, nil
}

func (c *coalitionService) GetMainMembers(uid int) []int {
	rids := make([]int, 0)
	coalition := c.GetCoalition(uid)
	if coalition != nil {
		chairman := coalition.Chairman
		viceChairman := coalition.ViceChairman
		rids = append(rids, chairman, viceChairman)
	}
	return rids
}
