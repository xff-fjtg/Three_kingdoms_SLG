package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"encoding/json"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

var RoleAttrService = &roleAttrService{
	attrs: make(map[int]*data.RoleAttribute),
}

type roleAttrService struct {
	mutex sync.RWMutex
	attrs map[int]*data.RoleAttribute
}

func (r *roleAttrService) Load() {
	var ras []*data.RoleAttribute
	err := global.DB.Find(&ras).Error
	if err != nil {
		log.Println("load RoleAttr fail")
		// 处理错误
		return
	}
	for _, v := range ras {
		r.attrs[v.RId] = v
	}
	//查询所有联盟，进行匹配
	l := CoalitionService.ListCoalition()
	for _, c := range l {
		for _, rid := range c.MemberArray {
			attr, ok := r.attrs[rid]
			if ok {
				attr.UnionId = c.Id
			}
		}
	}
}
func (r *roleAttrService) TryCreate(rid int, conn net.WSConn, tx *gorm.DB) error {
	role := &data.RoleAttribute{}
	result := global.DB.Where("rid = ?", rid).First(&role)
	//找角色
	if result.RowsAffected == 0 {
		role.RId = rid
		role.UnionId = 0
		role.ParentId = 0
		role.PosTags = ""
		role.LastCollectTime = time.Now()
		err := global.DB.Create(&role).Error
		if err != nil {
			log.Println("插入RoleAttribute fail", err)
			tx.Rollback()
			return common.New(utils.DBError, "插入RoleAttribute fail")
		}
		r.mutex.Lock()
		r.attrs[rid] = role
		r.mutex.Unlock()
	} else if result.Error != nil {
		log.Println("RoleAttribute find fail")
		tx.Rollback()
		return common.New(utils.DBError, "database wrong")
	} else {
		//r.mutex.Lock()
		//r.attrs[rid] = role
		//r.mutex.Unlock()
		return nil
	}
	return nil
}
func (r *roleAttrService) GetTagList(rid int) ([]model.PosTag, error) {
	// 初始化一个空的posTags切片
	posTags := make([]model.PosTag, 0)
	//先去缓存中找
	ra, ok := r.attrs[rid]
	if ok {
		// 获取查询到的角色属性记录中的posTags字段内容
		tags := ra.PosTags
		// 如果posTags字段不为空
		if tags != "" {
			// 将posTags字段内容解析为JSON格式，并存储到posTags切片中
			err := json.Unmarshal([]byte(tags), &posTags)
			if err != nil {
				// 如果解析JSON出错，返回错误码和错误信息
				return nil, common.New(utils.DBError, "Database wrong")
			}
		}
	}
	// r.attrs里面没有缓存 那么就去查找数据库
	// 初始化一个RoleAttribute结构体指针
	ra = &data.RoleAttribute{}
	// 在数据库中查找给定rid的RoleAttribute记录
	result := global.DB.Where("rid =?", rid).Find(&ra)
	// 检查数据库查询是否有错误发生
	if result.Error != nil {
		// 如果有错误，返回错误码和错误信息
		return nil, common.New(utils.DBError, "查找Tag fail")
	} else if result.RowsAffected == 0 {
		// 如果没有找到记录，返回错误码和错误信息
		return nil, common.New(utils.DBError, "查找Tag fail")
	} else {

	}
	// 返回posTags切片和nil错误信息，表示操作成功
	return posTags, nil
}

func (r *roleAttrService) Get(rid int) (*data.RoleAttribute, error) {
	ra, ok := r.attrs[rid]
	if ok {
		return ra, nil
	}
	return nil, nil
}

func (r *roleAttrService) GetUnion(rid int) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ra, ok := r.attrs[rid]
	if ok {
		return ra.UnionId
	}
	log.Println("GetUnion fail")
	return 0
}

func (r *roleAttrService) GetParentId(rid int) int {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ra, ok := r.attrs[rid]
	if ok {
		return ra.ParentId
	}
	log.Println("GetUnion fail")
	return 0
}
