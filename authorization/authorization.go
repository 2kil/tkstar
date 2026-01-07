/*
 * @Author: 2Kil
 * @Date: 2026-01-07 18:48:47
 * @LastEditors: 2Kil
 * @LastEditTime: 2026-01-07 18:49:12
 * @Description:简单授权
 */
package authorization

import (
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
	Code string
	Data []Accredit
	mu   sync.Mutex
}

// NewClient 创建一个新的客户端实例
func NewClient(code string) *Client {
	return &Client{Code: code}
}

// GetAccredit 获取授权信息，更新内部缓存并返回数据
func (c *Client) GetAccredit() ([]Accredit, error) {
	url := "https://active.clewm.net/" + c.Code
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	jumpURL := ""
	tableData := make(map[string]string)

	// 重试获取 jump_url 逻辑
	for i := 0; i < 2; i++ {
		resp, err := client.Do(req)
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
			resp, err := client.Do(req)
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

				// 提取表格行
				rows := reRow.FindAllStringSubmatch(tableHTML, -1)

				// 遍历每一行
				for _, row := range rows {
					// 提取单元格
					cells := reCell.FindAllStringSubmatch(row[1], -1)

					// 创建 map 存储单元格数据
					// 遍历每一列，这里假设列的顺序是固定的
					if len(cells) >= 2 {
						sn := removeHTMLTags(cells[0][1])
						time := removeHTMLTags(cells[1][1])
						tableData[sn] = time
					}
				}
				// fmt.Println("Table Data:", tableData)
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

// CheckAccredit 检查指定 key 的授权是否有效
func (c *Client) CheckAccredit(key string) bool {
	c.mu.Lock()
	data := c.Data
	c.mu.Unlock()

	if len(data) == 0 {
		var err error
		data, err = c.GetAccredit()
		if err != nil {
			return false
		}
	}

	var val string
	found := false
	for _, item := range data {
		if item.Sn == key {
			val = item.Time
			found = true
			break
		}
	}

	if !found {
		return false
	}

	// 尝试解析时间，支持常见格式
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, strings.TrimSpace(val), time.Local); err == nil {
			// 如果当前时间大于 value 则返回 false，否则返回 true
			return !time.Now().After(t)
		}
	}

	return false
}
