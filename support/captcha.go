package support

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/kataras/golog"
	"github.com/mojocn/base64Captcha"
	"math/rand"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
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

// PortCheckup 传入查询的端口号
// 返回端口号对应的进程PID，若没有找到相关进程，返回-1
func PortCheckup(portNumber int) (pid int, err error) {
	pid = -1
	err = nil
	var os = runtime.GOOS // 当前操作系统
	var (
		cmd      *exec.Cmd    // 命令控制台
		outBytes bytes.Buffer // 控制台响应
	)
	switch os {
	case "windows":
		cmd = exec.Command("cmd", "/c", fmt.Sprintf("netstat -ano -p tcp | findstr %d", portNumber))
	case "linux":
		cmd = exec.Command("bash", "-c", fmt.Sprintf("lsof -i:%d", portNumber))
	default: // 其他系统
		golog.Warn("当前系统：", os)
		return pid, errors.New(fmt.Sprintf("当前系统:%s,未支持！", os))
	}
	// 显示运行的命令
	golog.Info("cmd args: ", cmd.Args)
	cmd.Stdout = &outBytes
	if err = cmd.Run(); err != nil {
		if err.Error() == "exit status 1" {
			err = nil //表示控制台无内容退出
		}
		return pid, err
	}
	cmdResult := outBytes.String()
	// 控制台结果为空，代表端口未被占用
	if cmdResult == "" {
		return pid, nil
	}
	golog.Info("port: ", portNumber, ", checkup:\r\n", cmdResult)
	if os == "windows" {
		var regPort = "\\d+\\.\\d+\\.\\d+\\.\\d:(\\d+)"
		var regPid = "LISTENING\\s+(\\d+)"
		for _, line := range strings.Split(cmdResult, "\r\n") {
			if fport := regexp.MustCompile(regPort).FindAllStringSubmatch(line, 1); fport != nil {
				if fport[0][1] == strconv.Itoa(portNumber) {
					fpid := regexp.MustCompile(regPid).FindAllStringSubmatch(line, 1)
					pid, err = strconv.Atoi(fpid[0][1])
					break
				}
			}
		}
	}
	if os == "linux" {
		var regPid = ".*?\\s+(\\d+)"
		fpid := regexp.MustCompile(regPid).FindAllStringSubmatch(cmdResult, 1)
		pid, err = strconv.Atoi(fpid[0][1])
	}
	return pid, err
}
