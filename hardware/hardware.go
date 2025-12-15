/*
 * @Author: 2Kil
 * @Date: 2025-12-15 11:47:25
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 11:50:20
 * @Description: 硬件相关
 */
 
package hardware

import (
	"crypto/md5"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"syscall"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
)

// SysGetSerialKey 获取设备硬件特征码。
// 结合 MAC 地址、系统 UUID 和硬盘序列号生成唯一的简短机器码。
// return: 机器码
func SysGetSerialKey() string {
	// 1. 获取本机的MAC地址
	var mac string
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			// 排除回环接口、未启用接口、无硬件地址接口
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 || len(iface.HardwareAddr) == 0 {
				continue
			}
			mac = iface.HardwareAddr.String()
			break
		}
	}

	// 2. 获取系统UUID (Windows)
	var uuid string
	// 注意：wmic 在较新的 Windows 版本中可能被废弃，但在旧系统中可用
	cmd := exec.Command("wmic", "csproduct", "get", "UUID")
	uuidOut, err := cmd.Output()
	if err != nil {
		uuid = "BC2B8100-FD76-11EE-BE99-DA3F32D12700" // 默认值
	} else {
		uuid = string(uuidOut)
	}

	// 3. 获取硬盘串号 (主硬盘 Index=0)
	var diskSerial string
	cmd = exec.Command("wmic", "diskdrive", "where", "Index=0", "get", "SerialNumber")
	diskSerialOut, err := cmd.Output()
	if err != nil {
		diskSerial = "6479_A771_20C0_1EFF" // 默认值
	} else {
		// 清理 wmic 输出
		diskSerial = string(diskSerialOut)
		diskSerial = strings.Replace(diskSerial, "SerialNumber", "", 1)
		diskSerial = strings.TrimSpace(diskSerial)
	}

	// 4. 生成 MD5 摘要
	rawKey := mac + uuid + diskSerial
	hash := md5.Sum([]byte(rawKey))
	reg0 := strings.ToUpper(fmt.Sprintf("%x", hash))

	// 5. 字符混淆替换
	replacer := strings.NewReplacer(
		"O", "0",
		"o", "0",
		"L", "1",
		"l", "1",
		"I", "2",
		"i", "2",
	)
	reg0 = replacer.Replace(reg0)

	// 6. 截取并拼接生成最终机器码
	// 确保字符串长度足够，防止切片越界 (MD5 长度为 32)
	if len(reg0) < 14 {
		return reg0 // 理论上 MD5 不会小于 32，但做防御性编程
	}
	return reg0[8:11] + reg0[2:3] + reg0[12:14]
}

// 定义常用键位的映射表
var keyMap = map[string]int{
	// 字母 (A-Z)
	"A": 0x41, "B": 0x42, "C": 0x43, "D": 0x44, "E": 0x45,
	"F": 0x46, "G": 0x47, "H": 0x48, "I": 0x49, "J": 0x4A,
	"K": 0x4B, "L": 0x4C, "M": 0x4D, "N": 0x4E, "O": 0x4F,
	"P": 0x50, "Q": 0x51, "R": 0x52, "S": 0x53, "T": 0x54,
	"U": 0x55, "V": 0x56, "W": 0x57, "X": 0x58, "Y": 0x59, "Z": 0x5A,

	// 数字 (0-9)
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,

	// 功能键 (F1-F12)
	"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73, "F5": 0x74,
	"F6": 0x75, "F7": 0x76, "F8": 0x77, "F9": 0x78, "F10": 0x79,
	"F11": 0x7A, "F12": 0x7B,

	// 特殊控制键 (区分左右)
	"LCTRL": 0xA2, "RCTRL": 0xA3, // 左/右 Ctrl
	"LSHIFT": 0xA0, "RSHIFT": 0xA1, // 左/右 Shift
	"LALT": 0xA4, "RALT": 0xA5, // 左/右 Alt (Menu)

	// 通用控制键 (不分左右，如果只想检测任意 Ctrl 用这个)
	"CTRL": 0x11, "SHIFT": 0x10, "ALT": 0x12,

	// 其他常用键
	"SPACE": 0x20, "ENTER": 0x0D, "ESC": 0x1B,
	"TAB": 0x09, "BACKSPACE": 0x08,
	"UP": 0x26, "DOWN": 0x28, "LEFT": 0x25, "RIGHT": 0x27,
}

// 判断按键是否按下
func KeyIsPress(keyName string) bool {
	// 1. 将输入转为大写，防止大小写敏感问题 (比如输入 "a" 也能识别)
	upperName := strings.ToUpper(keyName)

	// 2. 从 Map 中查找对应的虚拟键码
	vKey, ok := keyMap[upperName]
	if !ok {
		// 如果没找到定义的键，默认返回 false，或者你可以选择 panic
		return false
	}

	// 3. 调用 Windows API
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vKey))

	// 4. 判断最高位 (0x8000) 是否为 1
	return (ret & 0x8000) != 0
}
