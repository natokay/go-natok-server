package dsmapper

import (
	"github.com/kataras/golog"
	"natok-server/model"
)

// GetUser 获取用户信息
func GetUser(user *model.NatokUser) *model.NatokUser {
	if ok, err := Engine.Get(user); !ok {
		if err != nil {
			golog.Error(err)
		}
		return nil
	}
	return user
}
