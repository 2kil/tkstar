/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-27 23:06:45
 * @Description:star
 */
package star

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
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
 * @description: 获取设备硬件码
 * @return {string} 硬件码
 */
func GetSerialKey() string {
	// 获取本机的MAC地址
	var mac string
	interfaces, err := net.Interfaces()
	if err != nil {
		mac = ""
	} else {
		mac = interfaces[0].HardwareAddr.String()
	}

	// 获取系统UUID
	var uuid string
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	uuidOut, err := cmd.Output()
	if err != nil {
		uuid = "FFFFFFFFF"
	}
	uuid = string(uuidOut)

	// 获取硬盘串号
	var diskSerial string
	cmd = exec.Command("wmic", "diskdrive", "get", "SerialNumber")
	diskSerialOut, err := cmd.Output()
	if err != nil {
		diskSerial = "6479_A771_20C0_1EFF"
	}
	diskSerial = string(diskSerialOut)

	reg0 := strings.ToUpper(fmt.Sprintf("%x", md5.Sum([]byte(mac+uuid+diskSerial))))

	// 简化设备码
	return reg0[8:11] + reg0[2:3] + reg0[12:14]
}

/**
 * @description: 错误检测
 * @param {error} err 错误信息
 * @param {string} errString 自定义错误提示
 * @return {*}
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
		log.Fatal(err)
	}

	// 获取修改时间
	modTime := fileInfo.ModTime()
	buildTime := modTime.Format("06.0102.1504")

	// 打印修改时间
	log.Printf("Build Time:%s", buildTime)
	return buildTime
}
