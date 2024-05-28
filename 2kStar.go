/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 17:49:00
 * @Description:star
 */
package star

import (
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

func IsBug() {
	log.Println("The current mode is debug")
}

/**
 * @description: 生成随机数
 * @param {int} max 最大值
 * @return {int} 随机数
 */
func RandAtomic(max int) int {
	var counter int64
	rand.New(rand.NewSource(time.Now().UnixNano() + atomic.AddInt64(&counter, 1)))
	randomNum := rand.Intn(max)
	return randomNum
}

/**
 * @description: 生成范围内的随机数
 * @param {int} min 最小值
 * @param {int} max 最大值
 * @return {int} 随机数
 */
func RandAtomicRadius(min, max int) int {
	if max <= min {
		return min
	}
	var counter int64
	rand.New(rand.NewSource(time.Now().UnixNano() + atomic.AddInt64(&counter, 1)))
	randomNum := min + rand.Intn(max-min+1)
	return randomNum
}

/**
 * @description: 错误检测
 * @param {error} err 错误信息
 * @param {string} errString 自定义错误提示
 * @return {bool} 无错误true 有错误false
 */
func CheckErr(err error, errString ...string) bool {
	errString = append(errString, "Error")
	if err != nil {
		log.Println(errString, err)
		return false
	}
	return true
}

/**
 * @description: 输出编译时间
 * @return {string} 当前文件编译时间
 */
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
	log.Printf("Build Time:%s", buildTime)
	return buildTime
}

/**
 * @description: 切片去重,去空
 * @param {[]string} 待处理的切片
 * @return {[]string} 处理后的切片
 */
func RemoveDuplicates(s []string) []string {
	m := make(map[string]bool)
	var result []string
	for _, item := range s {
		if item == "" {
			continue // 跳过空字符串
		}
		if _, ok := m[item]; !ok {
			m[item] = true
			result = append(result, item)
		}
	}
	return result
}
