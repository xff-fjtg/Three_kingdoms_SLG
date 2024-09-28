package web

import (
	"Three_kingdoms_SLG/core"
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/server/web/controller"
	"Three_kingdoms_SLG/server/web/middleware"
	"github.com/gin-gonic/gin"
)

func Init(router *gin.Engine) {
	global.DB = core.InitGorm()
	initRouter(router)
}

func initRouter(router *gin.Engine) {
	router.Any("/account/register", middleware.Cors(), controller.DefaultAccountController.Register)
}
