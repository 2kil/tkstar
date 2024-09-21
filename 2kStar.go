/*
 * @Author: 2Kil
 * @Date: 2024-04-19 10:54:20
 * @LastEditors: 2Kil
 * @LastEditTime: 2024-09-21 23:04:07
 * @Description:tktar
 */
package tkstar

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
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
 * @description:go的curl实现
 * @param {string} cUrl(bash)格式的请求命令
 * @return {*} 响应体
 */
func NetCurl(curlBash string) (int, string) {
	// 解析curl命令
	method, url, headers, data, err := NetParseCurlComd(curlBash)
	if err != nil {
		fmt.Println("Error parsing curl command:", err)
		return 0, ""
	}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return 0, ""
	}

	// 设置HTTP头
	req.Header = headers

	// 如果你的数据体是JSON，并且你还没有设置Content-Type头，你可以在这里添加它
	// req.Header.Set("Content-Type", "application/json")

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending HTTP request:", err)
		return 0, ""
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return 0, ""
	}
	return resp.StatusCode, string(body)
}

/**
 * @description:go的curl实现(走代理)
 * @param {string} cUrl(bash)格式的请求命令
 * @return {*} 响应体
 */
func NetProxyCurl(proxy, curlBash string) (int, string) {
	// 解析curl命令
	method, urll, headers, data, err := NetParseCurlComd(curlBash)
	if err != nil {
		fmt.Println("Error parsing curl command:", err)
		return 0, ""
	}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest(method, urll, bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error creating HTTP request:", err)
		return 0, ""
	}

	// 设置HTTP头
	req.Header = headers

	// 发送HTTP请求
	var client *http.Client
	if proxy != "" {
		// 解析代理URL
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			log.Fatalf("Failed to parse proxy URL: %v", err)
		}

		// 创建一个http.Transport，设置Proxy字段为之前解析的URL
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		// 创建一个http.Client，使用上面配置的Transport
		client = &http.Client{
			Transport: transport,
		}
	} else {
		client = &http.Client{}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending HTTP request:", err)
		return 0, ""
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return 0, ""
	}
	return resp.StatusCode, string(body)
}

/**
 * @description: 解析curl命令
 * @param {string} curlCmd
 * @return {*} 请求方法,请求地址,请求头,请求体,错误信息
 */
func NetParseCurlComd(curlCmd string) (string, string, http.Header, []byte, error) {
	method := "GET"
	// 提取URL
	urlRe := regexp.MustCompile(`curl '([^']+)'`)
	urlMatch := urlRe.FindStringSubmatch(curlCmd)
	if len(urlMatch) != 2 {
		return method, "", nil, nil, fmt.Errorf("failed to find URL in curl command")
	}
	rawURL := urlMatch[1]
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return method, "", nil, nil, err
	}

	// 提取HTTP头
	headerRe := regexp.MustCompile(`-H '([^']+): ([^']+)'`)
	headers := make(http.Header)
	matches := headerRe.FindAllStringSubmatch(curlCmd, -1)
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		key, value := match[1], match[2]
		headers.Set(key, value)
	}

	//提取请求方法
	methodRe := regexp.MustCompile(`--request ([A-Za-z]+)`)
	methodMatch := methodRe.FindStringSubmatch(curlCmd)
	if len(methodMatch) >= 2 {
		//获取到请求方法
		method = methodMatch[1]
	}

	// 提取提交数据体
	dataRe := regexp.MustCompile(`--data-raw '([^']+)'`)
	dataMatch := dataRe.FindStringSubmatch(curlCmd)
	if len(dataMatch) >= 2 {
		rawData := []byte(dataMatch[1])
		return method, parsedURL.String(), headers, rawData, nil
	}

	return method, parsedURL.String(), headers, nil, nil
}

/**
 * @description: 记录日志到文件,添加引用[log "github.com/sirupsen/logrus"]
 * @param {string} logFIle
 * @return {*}
 */
func LogFile(logFIle string) (*os.File, error) {
	// 创建一个文件用于写入日志
	logFile, err := os.OpenFile(logFIle, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("error opening file: %v", err)
		return logFile, err
	}
	// defer logFile.Close()
	log.SetOutput(logFile)
	return logFile, nil
}
