package logic

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/game/model"
	"Three_kingdoms_SLG/server/game/model/data"
	"Three_kingdoms_SLG/utils"
	"log"
)

type roleSkillService struct {
}

var RoleSkill = &roleSkillService{}

func (r *roleSkillService) GetSkill(rid int) ([]model.Skill, error) {
	roleSkill := make([]data.Skill, 0)
	err := global.DB.Where("rid = ?", rid).Find(&roleSkill).Error
	if err != nil {
		log.Println("search roleSkill fail")
		return nil, common.New(utils.DBError, "search roleSkill fail")
	}
	modelSkill := make([]model.Skill, 0)
	for _, v := range roleSkill {
		modelSkill = append(modelSkill, v.ToModel().(model.Skill))
	}
	return modelSkill, nil
}
