/*
 * @Author: 2Kil
 * @Date: 2025-09-28 13:46:43
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 10:24:52
 * @Description:网络相关
 */
package tkstar

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// NetProxyCurl 使用指定的代理执行一个类似于curl命令的网络请求。
// 它解析curl命令字符串，通过代理发送请求，并返回HTTP状态码和响应体。
func NetProxyCurl(proxy, curlBash string) (int, string, error) {
	// 解析curl命令
	method, urll, headers, data, err := NetParseCurlComd(curlBash)
	if err != nil {
		return 0, "", fmt.Errorf("解析curl命令时出错: %w", err)
	}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest(method, urll, bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("创建HTTP请求时出错: %w", err)
	}

	// 设置HTTP头
	req.Header = headers

	// 发送HTTP请求
	var client *http.Client
	if proxy != "" {
		// 解析代理URL
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			// 使用 log.Fatalf 会导致程序退出，这里改为返回错误
			return 0, "", fmt.Errorf("解析代理URL失败: %w", err)
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
		return 0, "", fmt.Errorf("发送HTTP请求时出错: %w", err)
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("读取响应体时出错: %w", err)
	}
	return resp.StatusCode, string(body), nil
}

// NetParseCurlComd 解析一个curl命令字符串，提取出请求方法、URL、请求头和请求体。
// 这个函数增强了对不同引号和参数格式的兼容性。
func NetParseCurlComd(curlCmd string) (string, string, http.Header, []byte, error) {
	method := "GET"
	headers := make(http.Header)
	var rawData []byte

	// 提取URL
	// 支持 'url', "url", 和无引号的url
	urlRe := regexp.MustCompile(`curl\s+(?:'([^']*)'|"([^"]*)"|([^\s-]+))`)
	urlMatch := urlRe.FindStringSubmatch(curlCmd)
	if len(urlMatch) < 2 {
		return "", "", nil, nil, fmt.Errorf("在curl命令中未找到URL")
	}
	rawURL := urlMatch[1]
	if rawURL == "" {
		rawURL = urlMatch[2]
	}
	if rawURL == "" {
		rawURL = urlMatch[3]
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", nil, nil, err
	}

	// 提取HTTP头
	// 支持 -H 或 --header, 以及 'key: value' 或 "key: value"
	headerRe := regexp.MustCompile(`(?:-H|--header)\s+'([^']*)'|"(.*?)"`)
	matches := headerRe.FindAllStringSubmatch(curlCmd, -1)
	for _, match := range matches {
		headerStr := match[1]
		if headerStr == "" && len(match) > 2 {
			headerStr = match[2]
		}
		parts := strings.SplitN(headerStr, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers.Set(key, value)
		}
	}

	//提取请求方法
	// 支持 -X 或 --request
	methodRe := regexp.MustCompile(`(?:-X|--request)\s+([A-Z]+)`)
	methodMatch := methodRe.FindStringSubmatch(curlCmd)
	if len(methodMatch) >= 2 {
		method = methodMatch[1]
	}

	// 提取提交数据体
	// 支持 --data-raw 或 --data, 以及 'data' 或 "data"
	dataRe := regexp.MustCompile(`(?:--data-raw|--data)\s+'([^']*)'|"(.*?)"`)
	dataMatch := dataRe.FindStringSubmatch(curlCmd)
	if len(dataMatch) >= 2 {
		dataStr := dataMatch[1]
		if dataStr == "" && len(dataMatch) > 2 {
			dataStr = dataMatch[2]
		}
		rawData = []byte(dataStr)
		// 如果有数据体，但未指定方法，通常默认为POST
		if len(methodMatch) == 0 {
			method = "POST"
		}
	}

	return method, parsedURL.String(), headers, rawData, nil
}

// NetCurl 执行一个类似于curl命令的网络请求。
// 它是 NetProxyCurl 的一个简化版本，不使用代理。
func NetCurl(curlBash string) (int, string, error) {
	// 解析curl命令
	method, urll, headers, data, err := NetParseCurlComd(curlBash)
	if err != nil {
		return 0, "", fmt.Errorf("解析curl命令时出错: %w", err)
	}

	// 创建一个HTTP POST请求
	req, err := http.NewRequest(method, urll, bytes.NewBuffer(data))
	if err != nil {
		return 0, "", fmt.Errorf("创建HTTP请求时出错: %w", err)
	}

	// 设置HTTP头
	req.Header = headers

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("发送HTTP请求时出错: %w", err)
	}
	defer resp.Body.Close()

	// 读取并打印响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("读取响应体时出错: %w", err)
	}
	return resp.StatusCode, string(body), nil
}
