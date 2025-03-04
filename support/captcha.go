package support

import (
	"github.com/mojocn/base64Captcha"
	"math/rand"
)

// 验证码信息常量
const (
	CaptchaId      = "NATOK-CAPTCHA"
	CaptchaWidth   = 120
	CaptchaHeight  = 44
	CaptchaLength  = 5
	CaptchaContent = "1234567890abcdefghijklmnopqrstwvxyzABCDEFGHIJKLMNOPQRSTWVXYZ"
	CaptchaFont    = "wqy-microhei.ttc"
)

// NewDriver 创建配置好的验证码驱动对象
func NewDriver() *base64Captcha.DriverString {
	driver := new(base64Captcha.DriverString)
	driver.Height = CaptchaHeight
	driver.Width = CaptchaWidth
	driver.NoiseCount = 4 + rand.Intn(5)
	driver.ShowLineOptions = base64Captcha.OptionShowHollowLine
	driver.Length = CaptchaLength
	driver.Source = CaptchaContent
	driver.Fonts = []string{CaptchaFont}
	return driver
}
