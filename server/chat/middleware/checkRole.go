package middleware

import (
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/utils"
)

func CheckRole() net.MiddlewareFunc {
	return func(next net.HandleFunc) net.HandleFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
			_, err := req.Conn.GetProperty("role")
			if err != nil {
				rsp.Body.Code = utils.SessionInvalid
				return
			}
			next(req, rsp)
		}
	}
}
