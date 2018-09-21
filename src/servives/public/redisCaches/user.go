package redisCaches

import (
	"core"
	. "core/libs"
	"encoding/json"
	"servives/login/model"
	"time"
)

const (
	DB_User_Key = "DB_User_"
)

//设置DBUser缓存
func SetDBUser(dbUser *model.DbUser) error {
	redisClient := core.Service.GetRedisClient("user")

	userKey := DB_User_Key + NumToString(dbUser.Id)
	userData, _ := json.Marshal(dbUser)
	return redisClient.Set(userKey, userData, time.Hour*24).Err()
}
