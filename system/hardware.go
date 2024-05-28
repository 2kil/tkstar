/*
 * @Author: 2Kil
 * @Date: 2024-05-28 11:19:31
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-05-28 11:20:47
 * @Description:硬件信息
 */
package star

import (
	"crypto/md5"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

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