// ----------------------------------------------
// @Author: 2Kil
// @Date: 2025-09-28 13:46:43
// @LastEditors: 2Kil
// @LastEditTime: 2025-09-28 13:47:13
// @Description: 网络相关
// ----------------------------------------------
package tkstar

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

// ----------------------------------------------
// @description: go的curl实现(走代理)
// @param {*} proxy 代理地址
// @param {string} curlBash
// @return {*}
// ----------------------------------------------
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

// ----------------------------------------------
// @description: 解析curl命令
// @param {string} curlCmd
// @return {*} 请求方法,请求地址,请求头,请求体,错误信息
// ----------------------------------------------
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

// ----------------------------------------------
// @description: go的curl实现
// @param {string} curlBash
// @return {*} 响应体
// ----------------------------------------------
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
