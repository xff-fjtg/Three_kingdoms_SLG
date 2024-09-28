package controller

import (
	"Three_kingdoms_SLG/server/common"
	"Three_kingdoms_SLG/server/web/logic"
	"Three_kingdoms_SLG/server/web/model"
	"Three_kingdoms_SLG/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

var DefaultAccountController = &AccountController{}

type AccountController struct {
}

func (a *AccountController) Register(c *gin.Context) {
	//1。获取请求参数
	//2.看看数据库有没有用户名，有就不注册，没有就注册
	rq := &model.RegisterReq{}
	// 在这种情况下，将自动选择合适的绑定
	err := c.ShouldBind(rq)
	fmt.Println(rq.Password)
	if err != nil {
		log.Println("参数格式不合法", err)
		c.JSON(http.StatusOK, common.Error(utils.InvalidParam, "参数不合法"))
		return
	}
	//一般web error会自定义
	err = logic.DefaultAccountLogic.Register(rq)
	if err != nil {
		log.Println("register fail", err)
		c.JSON(http.StatusOK, common.Error(err.(*common.MyError).Code(), err.Error()))
		return
	}
	c.JSON(http.StatusOK, common.Success(utils.OK, nil))
}
