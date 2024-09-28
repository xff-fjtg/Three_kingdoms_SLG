package net

import (
	"log"
	"strings"
	"sync"
)

type HandleFunc func(req *WsMsgReq, rsp *WsMsgRsp)

type MiddlewareFunc func(handleFunc HandleFunc) HandleFunc

type Group struct {
	prefix        string
	handlerMap    map[string]HandleFunc
	middlewareMap map[string][]MiddlewareFunc
	Middlewares   []MiddlewareFunc
	mutex         sync.RWMutex
}

func (g *Group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp) {
	h, ok := g.handlerMap[name]
	if !ok {
		h, ok = g.handlerMap["*"]
		if !ok {
			log.Println("not found router", name)
			return
		}
	}
	if ok {
		//中间件 执行路由之前的代码
		for i := 0; i < len(g.Middlewares); i++ { //全局
			h = g.Middlewares[i](h)
		}
		mm, ok := g.middlewareMap[name]
		if ok {
			for i := 0; i < len(mm); i++ { //个别
				h = mm[i](h)
			}
		}
		h(req, rsp)
	}
}
func (g *Group) AddRouter(name string, handleFunc HandleFunc, middlewares ...MiddlewareFunc) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.handlerMap[name] = handleFunc
	g.middlewareMap[name] = middlewares

}
func (g *Group) Use(middlewares ...MiddlewareFunc) { //全部一起用的
	g.Middlewares = append(g.Middlewares, middlewares...)
}

type Router struct {
	group []*Group
}

func (r *Router) Group(prefix string) *Group {
	g := &Group{
		prefix:        prefix,
		handlerMap:    make(map[string]HandleFunc),
		middlewareMap: make(map[string][]MiddlewareFunc),
		Middlewares:   make([]MiddlewareFunc, 0),
	}
	r.group = append(r.group, g)
	return g
}

func (r *Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	//req.Body.Name 路径 登陆业务 account.login login路由标识
	strs := strings.Split(req.Body.Name, ".")
	prefix := ""
	name := ""
	if len(strs) == 2 {
		prefix = strs[0]
		name = strs[1]
	} //prefix 是 "account"，表示消息的类别或模块
	//name 是 "login"，表示具体的操作
	for _, g := range r.group {
		if g.prefix == prefix {
			g.exec(name, req, rsp)
		} else if g.prefix == "*" {
			g.exec(name, req, rsp)
		}

	}
}
