package module

import (
	. "core/libs"
	"core/libs/random"
	"core/libs/sessions"
	"core/protos"
	"core/protos/gameProto"
	"github.com/golang/protobuf/proto"
	"servives/login/cache"
	"servives/public"
	"servives/public/dbModels"
	"servives/public/redisCaches"
)

//登录
func Login(clientSession *sessions.BackSession, msgData proto.Message) {
	data := msgData.(*gameProto.UserLoginC2S)
	account := data.GetAccount()

	onlineUser := cache.GetOnlineUserByAccount(account)
	if onlineUser != nil {
		oldClientSession := onlineUser.Session
		if oldClientSession.ID() != clientSession.ID() {
			//当前在线，但是连接不同，其他客户端连接，需通知当前客户端下线
			sendOtherLogin(oldClientSession)
			//替换Session
			cache.RemoveOnlineUser(oldClientSession.ID())
			//登录成功后处理
			loginSuccess(clientSession, onlineUser.Account, onlineUser.UserID)
		}
	} else {
		//进行DB登录
		dbUser := login(account)
		//登录成功后处理
		loginSuccess(clientSession, dbUser.Account, dbUser.Id)
	}
}

func login(account string) *dbModels.DbUser {
	//db中获取用户数据
	dbUser := dbModels.GetDbUser(account)
	if dbUser == nil {
		//注册
		addMoney := random.RandomInt31n(999)
		dbUser = dbModels.AddDbUser(account, addMoney)
	}
	//加入redis缓存
	redisCaches.SetDBUser(dbUser)
	return dbUser
}

//登录成功后处理
func loginSuccess(clientSession *sessions.BackSession, account string, userID uint64) {
	//缓存用户在线数据
	cache.AddOnlineUser(userID, account, clientSession)
	clientSession.AddCloseCallback(nil, "user.loginSuccess", func() {
		cache.RemoveOnlineUser(clientSession.ID())
		DEBUG("用户下线：当前在线人数", cache.GetOnlineUsersNum())
	})
	DEBUG("用户上线：当前在线人数", cache.GetOnlineUsersNum())

	//返回客户端数据
	token := public.CreateToken(NumToString(userID))
	sendMsg := &gameProto.UserLoginS2C{
		Token: protos.String(token),
	}
	public.SendMsgToClient(clientSession, sendMsg)
}

func sendOtherLogin(clientSession *sessions.BackSession) {
	sendMsg := &gameProto.UserOtherLoginNoticeS2C{}
	public.SendMsgToClient(clientSession, sendMsg)
}
