package middleware

import (
	"Three_kingdoms_SLG/net"
	"fmt"
	"log"
)

func Log() net.MiddlewareFunc {
	return func(handleFunc net.HandleFunc) net.HandleFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			log.Println("请求路由", req.Body.Name)
			log.Println("请求参数", fmt.Sprintf("%v", req.Body.Msg))
			handleFunc(req, rsp)
		}
	}
}
