package chat

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/chat/controller"
)

var Router = &net.Router{}

func Init() {

	initRouter()
}

func initRouter() {
	controller.ChatController.Router(Router)
}
