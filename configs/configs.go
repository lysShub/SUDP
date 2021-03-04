package configs

import (
	"os"
	"path/filepath"
)

// 全局常量变量

// GetRootPath 获取可执行二进制文件所在文件夹路径，不是代码文件所在路径
func GetRootPath() string {
	return filepath.ToSlash(filepath.Dir(os.Args[0])) + `/`
}

// ROOTPATH 可执行文件所在根目录
var ROOTPATH string = GetRootPath()

// LOGPATH 日志文件夹路径
var LOGPATH string = ROOTPATH + "/log/"

// ERRLOGPATH 错误日志路径
var ERRLOGPATH string = LOGPATH + "error.log"
