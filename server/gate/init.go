package gate

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/gate/controller"
)

var Router = &net.Router{}

func Init() {
	InitRouter()
}
func InitRouter() {
	controller.GateHandler.Router(Router)
}
