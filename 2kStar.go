/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-10-21 11:24:39
 * @Description:tktar
 */
package tkstar

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/logger"
)

/**
 * @description: 判断当前环境是否为Debug模式
 * @return {bool} true:Debug模式 false:Release模式
 */
func IsDebug() bool {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return false
	}
	if strings.Contains(dir, "Temp") && strings.Contains(dir, "go-build") {
		log.Println("The current mode is debug")
		return true
	}
	return false
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
func HelperRemoveDuplicates(s []string) []string {
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

/**
 * @description: aes加密
 * @param {string} 待加密的文本
 * @param {string} 16,24,32密钥
 * @return {string} 密文
 */
func TextAesEncrypt(plainText, key string) string {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	plainTextBytes := []byte(plainText)
	cipherText := make([]byte, aes.BlockSize+len(plainTextBytes))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(crand.Reader, iv); err != nil {
		return ""
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainTextBytes)
	text := base64.StdEncoding.EncodeToString(cipherText)

	//替换base64特殊字符
	text = strings.ReplaceAll(text, "/", "*")
	text = strings.ReplaceAll(text, "==", "#")
	text = strings.ReplaceAll(text, "=", "$")

	return text
}

/**
 * @description: aes解密
 * @param {string} 待解密的文本
 * @param {string} 密钥
 * @return {string} 明文
 */
func TextAesDecrypt(cipherText, key string) string {
	//替换base64特殊字符
	cipherText = strings.ReplaceAll(cipherText, "*", "/")
	cipherText = strings.ReplaceAll(cipherText, "$", "=")
	cipherText = strings.ReplaceAll(cipherText, "#", "==")

	cipherTextBytes, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return ""
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return ""
	}

	if len(cipherTextBytes) < aes.BlockSize {
		return ""
	}

	iv := cipherTextBytes[:aes.BlockSize]
	cipherTextBytes = cipherTextBytes[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherTextBytes, cipherTextBytes)

	return string(cipherTextBytes)
}

/**
 * @description: 获取设备硬件码
 * @return {string} 硬件码
 */
func SysGetSerialKey() string {
	// 获取本机的MAC地址
	var mac string
	interfaces, err := net.Interfaces()
	if err != nil {
		mac = ""
	} else {
		// 遍历所有网络接口，寻找一个合适的MAC地址
		for _, iface := range interfaces {
			// 排除回环接口 (loopback interface, e.g., "lo" or "Loopback Pseudo-Interface 1")
			// 排除未启用的接口 (interface not up)
			// 排除没有硬件地址的接口 (e.g., some virtual interfaces or software-only interfaces)
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 || len(iface.HardwareAddr) == 0 {
				continue
			}
			mac = iface.HardwareAddr.String()
			break // 找到第一个符合条件的MAC地址后即停止
		}
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
 * @description: 记录日志到文件&系统事件
 * @param {string} logFIle
 * @param {bool} systemLog true写入系统事件
 * @return {*}
 */
func LogFile(logFIle string, systemLog string) (*logger.Logger, error) {
	// 创建一个文件用于写入日志
	logFile, err := os.OpenFile(logFIle+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Fatalf("Failed to open log file: %v", err)
		return nil, err
	}
	// defer logFile.Close()

	loger := logger.Init(logFIle, true, true, logFile)
	// defer loger.Close()

	return loger, nil
}
