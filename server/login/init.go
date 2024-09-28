package login

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/login/controller"
)

var Router = &net.Router{}

func Init() {
	InitRouter()
}

func InitRouter() {
	controller.DefaultAccount.Router(Router)
}
