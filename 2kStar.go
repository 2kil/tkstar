/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-04-19 15:47:31
 * @Description:star
 */
package star

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"
)

var counter int64

func IsBug() {
	log.Println("hello")
}

/**
 * @description: 生成随机数
 * @param {int} max 最大值
 * @return {int} 随机数
 */
func RandAtomic(max int) int {
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
