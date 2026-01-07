/*
 * @Author: 2Kil
 * @Date: 2026-01-07 18:48:47
 * @LastEditors: 2Kil
 * @LastEditTime: 2026-01-07 22:37:53
 * @Description:简单授权
 */
package authorization

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 预编译正则以提升性能
var (
	reJumpURL = regexp.MustCompile(`var jump_url="(.*?)";`)
	reTable   = regexp.MustCompile(`(?s)(<table.*?</table>)`)
	reRow     = regexp.MustCompile(`(?s)<tr.*?>(.*?)</tr>`)
	reCell    = regexp.MustCompile(`(?s)<td.*?>(.*?)</td>`)
	reTags    = regexp.MustCompile(`<[^>]*>`)
)

// Accredit 导出结构体，包含授权信息的序列号和时间
type Accredit struct {
	Sn   string
	Time string
}

// Client 授权客户端，用于管理请求和缓存
type Client struct {
	Code       string //二维码网址
	Pwd        string //密码
	Data       []Accredit
	mu         sync.Mutex
	httpClient *http.Client
}

// NewClient 创建一个新的客户端实例
func NewClient(code string, pwd ...string) *Client {
	p := ""
	if len(pwd) > 0 {
		p = pwd[0]
	}
	return &Client{
		Code:       code,
		Pwd:        p,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// GetAccredit 获取授权信息，更新内部缓存并返回数据 旧版活码
// url格式 active.clewm.net/q8tDtnl
func (c *Client) GetAccredit() ([]Accredit, error) {
	url := "https://" + c.Code
	// 使用共享的 httpClient
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	jumpURL := ""
	tableData := make(map[string]string)

	// 重试获取 jump_url 逻辑
	for i := 0; i < 2; i++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			log.Println("Error during request:", err)
			time.Sleep(3 * time.Second) // 等待3秒
			continue
		}

		bodyText, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println("Error reading body:", err)
			time.Sleep(3 * time.Second) // 等待3秒
			continue
		}

		match := reJumpURL.FindStringSubmatch(string(bodyText))

		if len(match) > 1 {
			jumpURL = match[1]
			// fmt.Println("jump_url:", jumpURL)
			break // 找到 jump_url, 退出循环
		} else {
			log.Println("jump_url not found, retrying...")
			time.Sleep(3 * time.Second) // 等待3秒
		}
	}

	if jumpURL != "" {
		// 创建新的请求以应用 Header
		req, err := http.NewRequest("GET", jumpURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")

		// 重试获取 tableData 逻辑
		for i := 0; i < 2; i++ {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				log.Println("Error following jump URL:", err)
				time.Sleep(3 * time.Second)
				continue
			}

			bodyText, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				log.Println("Error reading body from jump URL:", err)
				time.Sleep(3 * time.Second)
				continue
			}

			match := reTable.FindStringSubmatch(string(bodyText))

			if len(match) > 1 {
				tableHTML := match[1]
				// 使用公共方法解析表格
				tableData = parseHTMLTable(tableHTML)
				break
			} else {
				log.Println("Table not found, retrying...")
				time.Sleep(3 * time.Second) // 等待3秒
			}
		}
	}

	if len(tableData) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data = make([]Accredit, 0, len(tableData))
	for k, v := range tableData {
		c.Data = append(c.Data, Accredit{Sn: k, Time: v})
	}
	return c.Data, nil
}

// removeHTMLTags 函数用于移除 HTML 标签
func removeHTMLTags(html string) string {
	return reTags.ReplaceAllString(html, "")
}

// parseHTMLTable 解析 HTML 表格数据
func parseHTMLTable(html string) map[string]string {
	data := make(map[string]string)
	rows := reRow.FindAllStringSubmatch(html, -1)
	for _, row := range rows {
		cells := reCell.FindAllStringSubmatch(row[1], -1)
		if len(cells) >= 2 {
			sn := removeHTMLTags(cells[0][1])
			t := removeHTMLTags(cells[1][1])
			data[sn] = t
		}
	}
	return data
}

// CheckAccredit 检查指定 key 的授权是否有效
func (c *Client) CheckAccredit(key string) bool {
	c.mu.Lock()
	data := c.Data
	c.mu.Unlock()

	if len(data) == 0 {
		var err error
		// 1. 如果配置了密码，优先尝试 GetAccredit2 (API方式)
		if c.Pwd != "" {
			data, err = c.GetAccredit2()
		}

		// 2. 如果没有数据（未配置密码或API失败），尝试 GetAccredit (网页解析方式)
		if len(data) == 0 {
			data, err = c.GetAccredit()
		}

		if err != nil {
			return false
		}
	}

	for _, item := range data {
		if item.Sn == key {
			return c.isTimeValid(item.Time)
		}
	}

	return false
}

// isTimeValid 辅助方法：验证时间字符串是否有效且未过期
func (c *Client) isTimeValid(val string) bool {
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, strings.TrimSpace(val), time.Local); err == nil {
			return !time.Now().After(t)
		}
	}
	return false
}

// GetAccredit2 获取授权信息，更新内部缓存并返回数据 新版活码
// url格式 qr61.cn/o78kxB/q8tDtnl
func (c *Client) GetAccredit2() ([]Accredit, error) {
	payload := fmt.Sprintf(`{
		"requests": [
			{
				"method": "POST",
				"timeout": 10000,
				"header": {
					"content-type": "application/x-www-form-urlencoded"
				},
				"path": "/qrcoderoute/qrcodeRouteNew",
				"body": "qrcode_route=%s&password=%s&render_default_fields=0&render_component_number=0&render_edit_btn=1&package_id="
			}
		]
	}`, c.Code, c.Pwd)
	var data = strings.NewReader(payload)
	req, err := http.NewRequest("POST", "https://nc.caoliao.net/batch-requests", data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("accept-language", "zh-CN,zh-TW;q=0.9,zh;q=0.8")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("dnt", "1")
	req.Header.Set("origin", "https://h5.clewm.net")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=1, i")
	req.Header.Set("referer", "https://h5.clewm.net/")
	req.Header.Set("sec-ch-ua", `"Microsoft Edge";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "cross-site")
	req.Header.Set("sec-fetch-storage-access", "active")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36 Edg/143.0.0.0")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 定义外层响应结构
	var result struct {
		Responses []struct {
			Body string `json:"body"`
		} `json:"responses"`
	}

	if err := json.Unmarshal(bodyText, &result); err != nil || len(result.Responses) == 0 {
		return nil, fmt.Errorf("failed to parse response")
	}

	// 解析内层 JSON 字符串
	var innerData struct {
		Data struct {
			QrcodeMsg struct {
				QrcodeComponent []struct {
					AttributeList []struct {
						ContentHtml struct {
							Value string `json:"value"`
						} `json:"content_html"`
					} `json:"attribute_list"`
				} `json:"qrcode_compontent"`
			} `json:"qrcode_msg"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(result.Responses[0].Body), &innerData); err != nil {
		return nil, err
	}
	tableData := make(map[string]string)
	if len(innerData.Data.QrcodeMsg.QrcodeComponent) > 0 && len(innerData.Data.QrcodeMsg.QrcodeComponent[0].AttributeList) > 0 {
		htmlValue := innerData.Data.QrcodeMsg.QrcodeComponent[0].AttributeList[0].ContentHtml.Value
		tableData = parseHTMLTable(htmlValue)
	}
	if len(tableData) == 0 {
		return nil, fmt.Errorf("no data found")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.Data = make([]Accredit, 0, len(tableData))
	for k, v := range tableData {
		c.Data = append(c.Data, Accredit{Sn: k, Time: v})
	}
	return c.Data, nil
}
