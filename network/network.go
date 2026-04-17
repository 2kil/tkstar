/*
 * @Author: 2Kil
 * @Date: 2025-09-28 13:46:43
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 11:44:30
 * @Description:网络相关
 */
package network

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	tokens, err := splitCurlCommand(curlCmd)
	if err != nil {
		return "", "", nil, nil, err
	}
	if len(tokens) == 0 || tokens[0] != "curl" {
		return "", "", nil, nil, fmt.Errorf("无效的curl命令")
	}

	var rawURL string
	explicitMethod := false

	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		switch token {
		case "-X", "--request":
			if i+1 >= len(tokens) {
				return "", "", nil, nil, fmt.Errorf("%s 缺少请求方法", token)
			}
			method = strings.ToUpper(tokens[i+1])
			explicitMethod = true
			i++
		case "-H", "--header":
			if i+1 >= len(tokens) {
				return "", "", nil, nil, fmt.Errorf("%s 缺少请求头内容", token)
			}
			parts := strings.SplitN(tokens[i+1], ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if key != "" {
					headers.Set(key, value)
				}
			}
			i++
		case "--data", "--data-raw", "-d":
			if i+1 >= len(tokens) {
				return "", "", nil, nil, fmt.Errorf("%s 缺少请求体内容", token)
			}
			rawData = []byte(tokens[i+1])
			if !explicitMethod {
				method = "POST"
			}
			i++
		case "--url":
			if i+1 >= len(tokens) {
				return "", "", nil, nil, fmt.Errorf("%s 缺少URL", token)
			}
			rawURL = tokens[i+1]
			i++
		default:
			if strings.HasPrefix(token, "-") {
				if skip, handled := curlOptionValueCount(token); handled {
					if i+skip >= len(tokens) {
						return "", "", nil, nil, fmt.Errorf("%s 缺少参数", token)
					}
					i += skip
				}
				continue
			}
			if rawURL == "" {
				rawURL = token
			}
		}
	}

	if rawURL == "" {
		return "", "", nil, nil, fmt.Errorf("在curl命令中未找到URL")
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", nil, nil, err
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

func splitCurlCommand(cmd string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		tokens = append(tokens, current.String())
		current.Reset()
	}

	for _, r := range cmd {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\' && !inSingle:
			escaped = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && !inSingle && !inDouble:
			flush()
		default:
			current.WriteRune(r)
		}
	}

	if escaped || inSingle || inDouble {
		return nil, fmt.Errorf("curl命令引号或转义不完整")
	}

	flush()
	return tokens, nil
}

func curlOptionValueCount(token string) (int, bool) {
	if strings.Contains(token, "=") {
		return 0, true
	}

	if strings.HasPrefix(token, "--") {
		switch token {
		case "--proxy", "--user", "--cookie", "--referer", "--compressed", "--connect-timeout", "--max-time", "--retry", "--cacert", "--cert", "--key", "--url", "--output", "--request-target":
			return 1, true
		}
		return 0, false
	}

	if len(token) == 2 && token[0] == '-' {
		switch token[1] {
		case 'x', 'U', 'u', 'b', 'e', 'm', 'o', 'E':
			return 1, true
		}
		return 0, false
	}

	if len(token) > 2 && token[0] == '-' && token[1] != '-' {
		switch token[1] {
		case 'x', 'U', 'u', 'b', 'e', 'm', 'o', 'E':
			return 0, true
		}
	}

	return 0, false
}
