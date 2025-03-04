package support

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GetCurrentAbPath 最终方案-全兼容
func GetCurrentAbPath() string {
	dir := getCurrentAbPathByExecutable()
	if strings.Contains(dir, getTmpDir()) {
		dir = getCurrentAbPathByCaller()
	}
	return dir + "/"
}

// 获取系统临时目录，兼容go run
func getTmpDir() string {
	dir := os.Getenv("TEMP")
	if dir == "" {
		dir = os.Getenv("TMP")
	}
	res, _ := filepath.EvalSymlinks(dir)
	return res
}

// 获取当前执行文件绝对路径
func getCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		logrus.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

// 获取当前执行文件绝对路径（go run）
func getCurrentAbPathByCaller() string {
	var abPath string
	if _, filename, _, ok := runtime.Caller(0); ok {
		if lst := strings.LastIndex(filename, "/support"); lst != -1 {
			abPath = filename[0:lst]
		}
	}
	return abPath
}
