package controller

import (
	"Three_kingdoms_SLG/global"
	"Three_kingdoms_SLG/net"
	"Three_kingdoms_SLG/server/login/model"
	"Three_kingdoms_SLG/server/login/proto"
	"Three_kingdoms_SLG/server/models"
	"Three_kingdoms_SLG/utils"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
)

var DefaultAccount = &Account{}

type Account struct {
}

func (a *Account) Router(r *net.Router) {
	g := r.Group("account")
	g.AddRouter("login", a.login)
	g.AddRouter("logout", a.logout)
	g.AddRouter("reLogin", a.reLogin) //重新登陆 比如断线了
}

func (a *Account) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//1.用户名 密码 硬件id
	//2.数据库匹配 查询user 匹配密码
	//3.保存用户登陆记录，最后一次登陆的信息
	//4.使用session，jwt
	//5.客户端发起登陆的时候可以这样判断是否合法
	loginReq := &proto.LoginReq{}
	loginRes := &proto.LoginRsp{}
	err := mapstructure.Decode(req.Body.Msg, loginReq)
	if err != nil {
		log.Println("change fail", err)
		return
	}
	username := loginReq.Username
	user := &models.User{}

	result := global.DB.Where("username = ? ", username).First(&user)
	if result.Error != nil {
		//没有这个用户
		rsp.Body.Code = utils.UserNotExist
		return
	}
	pwd := utils.Password(loginReq.Password, user.Passcode) //传入的 自己的

	if pwd != user.Passwd {
		//密码不对
		rsp.Body.Code = utils.PwdIncorrect
		return
	}
	//jwt A.B.C 三部分 A定义加密算法 B定义放入的数据 C部分 根据秘钥+A和B生成加密字符串
	token, _ := utils.Award(user.UId)
	rsp.Body.Code = utils.OK
	loginRes.UId = user.UId
	loginRes.Username = user.Username
	loginRes.Session = token
	loginRes.Password = ""
	rsp.Body.Msg = loginRes
	//3.保存用户登陆记录
	loginHistory := &model.LoginHistory{
		UId: user.UId, CTime: time.Now(), Ip: loginReq.Ip,
		Hardware: loginReq.Hardware, State: model.Login,
	}
	err = global.DB.Create(&loginHistory).Error
	if err != nil {
		log.Fatal("save user login history fail", err)
	}
	//最后一次登陆的信息
	logoutTime := time.Now()
	lastLogin := &model.LoginLast{}
	count := global.DB.Where("uid = ?", user.UId).First(&lastLogin)
	if count.Error != nil {
		//有数据 更新
		lastLogin.IsLogout = 0
		lastLogin.Ip = loginReq.Ip
		lastLogin.LoginTime = &logoutTime
		lastLogin.Session = token
		lastLogin.Hardware = loginReq.Hardware
		lastLogin.UId = user.UId
		err = global.DB.Save(&lastLogin).Error
		if err != nil {
			log.Fatal("save user last login history fail", err)
		}
	} else {
		lastLogin.IsLogout = 0
		lastLogin.Ip = loginReq.Ip
		lastLogin.LoginTime = &logoutTime
		lastLogin.Session = token
		lastLogin.LogoutTime = nil
		lastLogin.Hardware = loginReq.Hardware
		lastLogin.UId = user.UId
		err = global.DB.Save(&lastLogin).Error
		if err != nil {
			log.Fatal("save user last login history fail", err)
		}
	}
	//缓存 用户和当前的ws链接
	//相当于 用户在别的地方登陆了，我要给你当前的断掉
	net.Mgr.UserLogin(req.Conn, user.UId, token)
}

func (a *Account) logout(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.LogoutReq{}
	rspObj := &proto.LogoutRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.UId = reqObj.UId
	rsp.Body.Code = utils.OK

	tt := time.Now()
	//登出，写记录
	lh := &model.LoginHistory{UId: reqObj.UId, CTime: tt, State: model.Logout}
	global.DB.Create(&lh)

	ll := &model.LoginLast{}
	ok := global.DB.Where("uid=?", reqObj.UId).Find(&ll)

	if ok.Error == nil && ok.RowsAffected > 0 {
		ll.IsLogout = 1
		ll.LogoutTime = &tt
		global.DB.Model(&ll).Select("is_logout", "logout_time").Updates(&ll)

	} else {
		ll = &model.LoginLast{UId: reqObj.UId, LogoutTime: &tt, IsLogout: 0}
		global.DB.Create(&ll)
	}
	//登出
	net.Mgr.UserLogout(req.Conn)
}

func (a *Account) reLogin(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ReLoginReq{}
	rspObj := &proto.ReLoginRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	if reqObj.Session == "" {
		rsp.Body.Code = utils.SessionInvalid
		return
	}

	rsp.Body.Msg = rspObj
	rspObj.Session = reqObj.Session
	//判断参数是否合法
	_, c, err := utils.ParseToken(reqObj.Session)
	if err != nil {
		rsp.Body.Code = utils.SessionInvalid
	} else {
		//数据库验证一下
		ll := &model.LoginLast{}
		global.DB.Where("uid=?", c.Uid).Find(&ll)

		if ll.Session == reqObj.Session {
			if ll.Hardware == reqObj.Hardware {
				rsp.Body.Code = utils.OK
				//对就重新登陆
				net.Mgr.UserLogin(req.Conn, c.Uid, reqObj.Session)
			} else {
				rsp.Body.Code = utils.HardwareIncorrect
			}
		} else {
			rsp.Body.Code = utils.SessionInvalid
		}
	}
}
