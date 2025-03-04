package dsmapper

import (
	"github.com/sirupsen/logrus"
	"natok-server/model"
)

// GetUser 获取用户信息
func GetUser(user *model.NatokUser) *model.NatokUser {
	if ok, err := Engine.Get(user); !ok {
		if err != nil {
			logrus.Errorf("%v", err.Error())
		}
		return nil
	}
	return user
}

// ChangePassword  修改密码
func ChangePassword(userId int64, oldPassword, newPassword string) bool {
	user := GetUser(&model.NatokUser{Id: userId, Password: oldPassword})
	if user != nil {
		user.Password = newPassword
		if _, err := Engine.Id(userId).Cols("password").Update(user); err == nil {
			return true
		} else {
			logrus.Errorf("%v", err.Error())
		}
	}
	return false
}
