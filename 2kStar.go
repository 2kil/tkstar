/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 10:47:05
 * @Description: tkstar utils
 */

package tkstar

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/logger"
)

// IsDebug 判断当前环境是否为 Debug 模式。
// 通过检查执行文件路径中是否包含 "Temp" 和 "go-build" 来判断（通常适用于 go run）。
// return: true: Debug模式, false: Release模式
func IsDebug() bool {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return false
	}
	// Windows 下 go run 生成的临时目录特征
	if strings.Contains(dir, "Temp") && strings.Contains(dir, "go-build") {
		log.Println("The current mode is debug")
		return true
	}
	return false
}

// RandAtomic 生成 [0, max) 范围内的随机整数。
// param: max 最大值（不包含）
// return: 随机数
func RandAtomic(max int) int {
	if max <= 0 {
		return 0
	}
	return rand.Intn(max)
}

// RandAtomicRadius 生成 [min, max] 范围内的随机整数。
// param: min 最小值
// param: max 最大值
// return: 随机数
func RandAtomicRadius(min, max int) int {
	if max < min {
		return min
	}
	// rand.Intn(n) 返回 [0, n)，所以需要 +1 来包含 max
	return min + rand.Intn(max-min+1)
}

// CheckErr 错误检测辅助函数。
// 如果 err 不为 nil，打印日志并返回 false；否则返回 true。
// param: err 错误信息
// param: errString 自定义错误提示前缀
// return: 无错误 true, 有错误 false
func CheckErr(err error, errString ...string) bool {
	if err != nil {
		prefix := "Error"
		if len(errString) > 0 {
			prefix = strings.Join(errString, " ")
		}
		log.Printf("%s: %v", prefix, err)
		return false
	}
	return true
}

// BuildTime 输出并返回当前可执行文件的最后修改时间。
// 注意：这并不是准确的编译时间，而是文件的最后写入时间。
// return: 时间字符串，格式 "06.0102.1504"
func BuildTime() string {
	// 获取当前程序的文件信息
	fileInfo, err := os.Stat(os.Args[0])
	if err != nil {
		return "0.0.0.0"
	}

	// 获取修改时间
	modTime := fileInfo.ModTime()
	buildTime := modTime.Format("06.0102.1504")

	// 打印修改时间
	log.Printf("Build Time: %s", buildTime)
	return buildTime
}

// HelperRemoveDuplicates 对字符串切片进行去重和去空处理。
// param: s 待处理的切片
// return: 处理后的切片
func HelperRemoveDuplicates(s []string) []string {
	m := make(map[string]struct{}) // 使用空结构体节省内存
	var result []string
	for _, item := range s {
		if item == "" {
			continue // 跳过空字符串
		}
		if _, ok := m[item]; !ok {
			m[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}



// LogFile 初始化日志系统，支持写入文件和系统日志。
// param: logFileName 日志文件名前缀 (不含 .log)
// param: systemLog 是否写入系统事件日志 (Windows Event Log / Syslog)
// return: Logger 实例指针, 错误信息
func LogFile(logFileName string, systemLog bool) (*logger.Logger, error) {
	// 创建一个文件用于写入日志
	// 0660: 所有者和组用户可读写
	file, err := os.OpenFile(logFileName+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
		return nil, err
	}

	// logger.Init 参数: (name, verbose, systemLog, logFile)
	// verbose=true: 输出到 stderr
	// systemLog: 是否写入系统日志
	l := logger.Init(logFileName, true, systemLog, file)

	return l, nil
}
