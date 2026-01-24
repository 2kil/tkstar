/*
 * @Author: 2Kil
 * @Date: 2026-01-24 23:21:52
 * @LastEditors: 2Kil
 * @LastEditTime: 2026-01-24 23:23:12
 * @Description:浏览器相关
 */

package tkEdge

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/storage"
	"github.com/chromedp/chromedp"
	"github.com/lxn/win"
)

type dataStore struct {
	mu                sync.RWMutex
	latestReqHead     network.Headers
	latestResHead     network.Headers
	latestQueryParams url.Values
	latestReqID       network.RequestID
}

var (
	store = &dataStore{}
	// StatusChan 用于外部接收浏览器关闭信号
	StatusChan = make(chan error, 1)
	// 用于接收导航指令的管道
	navChan = make(chan string, 1)
	// 保存当前的浏览器上下文，供 GetCookies 等函数使用
	browserCtx context.Context
	// 保存当前上下文的取消函数，用于 Stop()
	browserCancel context.CancelFunc

	relativePath = "./edge_user_data"
	absPath, _   = filepath.Abs(relativePath)
)

// GetRes 获取响应头信息
// 获取指定路径的响应头信息(大小写敏感)
func GetRes(key string) string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if v, ok := store.latestResHead[key]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// GetReq 获取请求头信息
// 获取请求头中的信息(忽略大小写)
func GetReq(key string) string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	// fmt.Println(store.latestReqHead)
	// 1. 直接尝试获取
	if v, ok := store.latestReqHead[key]; ok {
		return fmt.Sprintf("%v", v)
	}

	// 2. 遍历查找（忽略大小写）
	targetKey := strings.ToLower(key)
	for k, v := range store.latestReqHead {
		if strings.ToLower(k) == targetKey {
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// 获取URL中的参数
// 获取GET请求中的urlQuery
func GetUrlQuery(key string) string {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.latestQueryParams.Get(key)
}

// Clear 清除存储的请求数据
func Clear() {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.latestReqHead = nil
	store.latestResHead = nil
	store.latestQueryParams = nil
	store.latestReqID = ""
}

// LoadUrl 加载指定url
func LoadUrl(targetURL string) {
	// 非阻塞发送，避免如果没有启动浏览器时卡死
	select {
	case navChan <- targetURL:
	default:
		fmt.Println("Warning: Browser not running or channel full")
	}
}

// GetUrl获取浏览器当前url
func GetUrl() (string, error) {
	if browserCtx == nil {
		return "", fmt.Errorf("浏览器尚未启动")
	}

	var currentURL string
	// 使用 chromedp.Location 获取当前页面的 URL
	err := chromedp.Run(browserCtx, chromedp.Location(&currentURL))
	if err != nil {
		return "", err
	}
	return currentURL, nil
}

// Stop 停止运行
func Stop() {
	if browserCancel != nil {
		browserCancel()
	}
}

// GetCookies 获取当前页面的所有 Cookie 并返回 Map 格式 [Name]Value
func GetCookies() (map[string]string, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("浏览器尚未启动")
	}

	var cookies []*network.Cookie
	err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))

	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	for _, c := range cookies {
		res[c.Name] = c.Value
	}
	return res, nil
}

// GetCookiesAll 获取所有 Cookie (Storage) 并返回 Map 格式 [Name]Value
func GetCookiesAll() (map[string]string, error) {
	if browserCtx == nil {
		return nil, fmt.Errorf("浏览器尚未启动")
	}

	var cookies []*network.Cookie
	err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = storage.GetCookies().Do(ctx)
		return err
	}))

	if err != nil {
		return nil, err
	}

	res := make(map[string]string)
	for _, c := range cookies {
		res[c.Name] = c.Value
	}
	return res, nil
}

// Run 有界面模式运行
// msedgePath可指定msedge.exe路径
func Run(urlPath string, msedgePath ...string) {
	//默认edge路径
	execPath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	if len(msedgePath) > 0 {
		execPath = msedgePath[0]
	}
	go func() {
		// 计算屏幕居中位置
		screenWidth := int(win.GetSystemMetrics(win.SM_CXSCREEN))
		screenHeight := int(win.GetSystemMetrics(win.SM_CYSCREEN))
		x := (screenWidth - 1400) / 2
		y := (screenHeight - 900) / 2

		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("headless", false),                      // 有界面
			chromedp.Flag("disable-gpu", false),                   // 启用显卡
			chromedp.Flag("enable-automation", false),             // 隐藏自动程序控制条
			chromedp.Flag("disable-extensions", false),            // 启用插件
			chromedp.Flag("disable-session-crashed-bubble", true), //禁用会话崩溃提示框
			chromedp.Flag("hide-crash-restore-bubble", true),      //隐藏崩溃恢复气泡
			chromedp.ExecPath(execPath),
			chromedp.WindowSize(1400, 900),
			chromedp.Flag("app", "about:blank"), // APP模式
			chromedp.Flag("window-position", fmt.Sprintf("%d,%d", x, y)),
			chromedp.UserDataDir(absPath),
		)

		err := commonRun(opts, urlPath)
		StatusChan <- err
	}()
}

// RunCli 无头模式运行 (Headless)
// msedgePath指定msedge.exe路径
func RunCli(urlPath string, msedgePath ...string) {
	//默认edge路径
	execPath := `C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe`
	if len(msedgePath) > 0 {
		execPath = msedgePath[0]
	}
	go func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.Flag("headless", true), // 无界面
			// 无头模式下通常建议禁用 GPU，除非特定场景需要
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("enable-automation", false),
			// 设置 User-Agent 伪装成正常浏览器，防止被无头检测拦截
			chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0"),
			chromedp.ExecPath(execPath),
			chromedp.WindowSize(1920, 1080), // 即使不可见也建议设置分辨率以确保渲染正确
			chromedp.UserDataDir(absPath),
		)

		err := commonRun(opts, urlPath)
		StatusChan <- err
	}()
}

// commonRun 抽取公共的启动和监听循环逻辑
func commonRun(opts []chromedp.ExecAllocatorOption, urlPath string) error {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// 创建上下文
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 赋值给全局变量
	browserCtx = ctx
	browserCancel = cancel

	// 监听逻辑
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *network.EventRequestWillBeSent:
			if strings.Contains(ev.Request.URL, urlPath) {
				u, _ := url.Parse(ev.Request.URL)
				store.mu.Lock()
				store.latestReqHead = ev.Request.Headers
				store.latestQueryParams = u.Query()

				// 修正：字段名为 RequestID
				store.latestReqID = ev.RequestID
				store.mu.Unlock()
			}

		case *network.EventRequestWillBeSentExtraInfo:
			store.mu.Lock()
			// 修正：字段名为 RequestID，类型为 network.RequestID
			if store.latestReqID == ev.RequestID {
				if store.latestReqHead == nil {
					store.latestReqHead = make(network.Headers)
				}
				// 合并 headers
				for k, v := range ev.Headers {
					store.latestReqHead[k] = v
				}

				// 处理关联的 Cookie (如果有)
				// 注意：ExtraInfo 里的 Cookie 是为了告知哪些被发送了，哪些被阻塞了
				// 如果你需要把它们拼成 Cookie 头字符串：
				if len(ev.AssociatedCookies) > 0 {
					var cookieStrs []string
					for _, c := range ev.AssociatedCookies {
						// 排除被阻塞的 cookie (可选逻辑)
						// if len(c.BlockedReasons) == 0 {
						cookieStrs = append(cookieStrs, fmt.Sprintf("%s=%s", c.Cookie.Name, c.Cookie.Value))
						// }
					}
					if len(cookieStrs) > 0 {
						// 如果原 headers 里没有 Cookie，或者你想追加
						// 这里简单处理为覆盖或设置
						store.latestReqHead["Cookie"] = strings.Join(cookieStrs, "; ")
					}
				}
			}
			store.mu.Unlock()

		case *network.EventResponseReceived:
			if strings.Contains(ev.Response.URL, urlPath) {
				store.mu.Lock()
				if store.latestResHead == nil {
					store.latestResHead = make(network.Headers)
				}
				for k, v := range ev.Response.Headers {
					store.latestResHead[k] = v
				}
				store.latestResHead["status_code"] = ev.Response.Status
				store.mu.Unlock()
			}
		}
	})

	// 初始导航到默认页面（比如抖音）
	if err := chromedp.Run(ctx, network.Enable(), chromedp.Navigate("https://www.douyin.com/?recommend=1")); err != nil {
		return err
	}

	// 循环处理信号
	for {
		select {
		case <-ctx.Done():
			// 浏览器被关闭（无论是外部 Stop 还是意外关闭）
			return fmt.Errorf("browser_closed")

		case targetURL := <-navChan:
			// 响应 LoadURL 函数发来的指令
			fmt.Printf("Navigating to: %s\n", targetURL)
			err := chromedp.Run(ctx, chromedp.Navigate(targetURL))
			if err != nil {
				fmt.Printf("Navigation failed: %v\n", err)
			}
		}
	}
}
