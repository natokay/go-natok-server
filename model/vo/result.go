package vo

import (
	"github.com/kataras/iris/v12/mvc"
	"github.com/sirupsen/logrus"
)

// Result struct 封装请求返回值
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

var (
	msgOk      = "ok"
	codeOk     = 20000
	codeFailed = 21000
)

func result(code int, msg string, data interface{}) *Result {
	return &Result{code, msg, data}
}

// GenSuccessData 数据封装
func GenSuccessData(data interface{}) *Result {
	return result(codeOk, msgOk, data)
}

// GenSuccessMsg 消息提示
func GenSuccessMsg(msg string) *Result {
	return result(codeOk, msg, nil)
}

// GenFailedMsg 失败提示
func GenFailedMsg(errMsg string) *Result {
	return result(codeFailed, errMsg, nil)
}

// TipMsg 若错误，将返回消息提示
func TipMsg(err error) mvc.Result {
	if err != nil {
		logrus.Errorf("%v", err.Error())
		return mvc.Response{Object: GenFailedMsg(err.Error())}
	}
	return mvc.Response{Object: GenSuccessMsg(msgOk)}
}

// TipErrorMsg 返回错误消息提示
func TipErrorMsg(errMsg string) mvc.Result {
	return mvc.Response{Object: GenFailedMsg(errMsg)}
}

// TipResult 返回数据
func TipResult(data interface{}) mvc.Result {
	return mvc.Response{Object: GenSuccessData(data)}
}
