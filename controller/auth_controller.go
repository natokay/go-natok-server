package controller

import (
	"encoding/json"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"github.com/mojocn/base64Captcha"
	"github.com/sirupsen/logrus"
	"natok-server/dsmapper"
	"natok-server/model"
	"natok-server/model/vo"
	"natok-server/support"
	"strings"
)

// AuthController struct 用户认证 - 控制层
type AuthController struct {
	Ctx     iris.Context
	Session *sessions.Session
}

func (c *AuthController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/verifyCode", "VerifyCodeHandler")
	b.Handle("POST", "/user/login", "Login")
	b.Handle("POST", "/user/logout", "Logout")
	b.Handle("GET", "/user/info", "UserInfo")
	b.Handle("POST", "/user/chgPwd", "ChangePassword")
}

// VerifyCodeHandler 生成图形验证码
func (c *AuthController) VerifyCodeHandler() {
	driver := support.NewDriver().ConvertFonts()
	newCaptcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
	_, content, answer := newCaptcha.Driver.GenerateIdQuestionAnswer()
	item, _ := newCaptcha.Driver.DrawCaptcha(content)
	item.WriteTo(c.Ctx.ResponseWriter())
	c.Session.Set(support.CaptchaId, answer)
	return
}

// Login 登录：设置用户session
func (c *AuthController) Login() mvc.Result {
	// 登录信息认证
	user := new(model.NatokUser)
	_ = c.Ctx.ReadJSON(user)

	// TODO 验证码非正常情况
	if code := c.Session.Get(support.CaptchaId); code == nil || code == "" {
		return vo.TipErrorMsg("当前不支持跨域；先关闭验证码再试试！")
	} else if len(user.Code) == 0 {
		return vo.TipErrorMsg("请输入验证码！")
	} else if strings.ToLower(code.(string)) != strings.ToLower(user.Code) {
		return vo.TipErrorMsg("验证码输入错误！")
	}

	if nil == dsmapper.GetUser(user) {
		return vo.TipErrorMsg("账号或密码错误！")
	}

	session := support.SessionsManager.Start(c.Ctx)
	session.Set(support.SessionKey, "OK")
	session.Set(support.SessionUserId, user.Id)

	marshal, _ := json.Marshal(user)
	logrus.Println(string(marshal))
	return vo.TipResult(map[string]string{"token": "natok-token"})
}

// Logout 登出：删除用户session
func (c *AuthController) Logout() mvc.Result {
	session := support.SessionsManager.Start(c.Ctx)
	session.Clear()
	return vo.TipResult("success")
}

// UserInfo 用户信息
func (c *AuthController) UserInfo() mvc.Result {
	ret := map[string]interface{}{
		"introduction": "Natok Manager",
		"avatar":       "/static/img/avatar.png",
		"name":         "Natok",
		"roles":        []string{"admin"},
	}
	return vo.TipResult(ret)
}

// ChangePassword 修改密码
func (c *AuthController) ChangePassword() mvc.Result {
	session := support.SessionsManager.Start(c.Ctx)
	if userId, err := session.GetInt64(support.SessionUserId); err == nil {
		oldPassword := c.Ctx.URLParam("oldPassword")
		newPassword := c.Ctx.URLParam("newPassword")
		if oldPassword == "" || newPassword == "" {
			return vo.TipErrorMsg("请输入原密码和新密码！")
		}
		if dsmapper.ChangePassword(userId, oldPassword, newPassword) {
			return vo.TipResult("success")
		} else {
			return vo.TipErrorMsg("原密码错误！")
		}
	} else {
		return vo.TipErrorMsg("请先登录！")
	}
}
