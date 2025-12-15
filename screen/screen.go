/*
 * @Author: 2Kil
 * @Date: 2025-12-15 11:22:19
 * @LastEditors: 2Kil
 * @LastEditTime: 2025-12-15 11:24:33
 * @Description: 屏幕相关
 */

package screen

import (
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// --- Windows API 常量定义 ---
// 这些常量对应 Windows 头文件中的定义，用于控制窗口样式、消息类型和绘图选项
const (
	// 窗口样式
	WS_POPUP         = 0x80000000 // 弹出式窗口（无标题栏、无边框）
	WS_VISIBLE       = 0x10000000 // 创建时即可见
	WS_EX_TOPMOST    = 0x00000008 // 扩展样式：总在最前
	WS_EX_TOOLWINDOW = 0x00000080 // 扩展样式：工具窗口（不在任务栏显示，Alt+Tab中不可见）
	WS_EX_LAYERED    = 0x00080000 // 扩展样式：分层窗口（用于实现透明效果）
	LWA_COLORKEY     = 0x00000001 // 分层属性：使用颜色键抠图（指定颜色变透明）

	// 窗口消息
	WM_PAINT   = 0x000F // 绘图消息
	WM_DESTROY = 0x0002 // 销毁消息

	// 类样式
	CS_HREDRAW = 0x0002 // 水平尺寸变化时重绘
	CS_VREDRAW = 0x0001 // 垂直尺寸变化时重绘

	// 文本绘制格式
	DT_CENTER     = 0x00000001 // 水平居中
	DT_VCENTER    = 0x00000004 // 垂直居中
	DT_SINGLELINE = 0x00000020 // 单行显示

	// 背景模式
	TRANSPARENT = 1 // 文本背景透明（即文字后面不画矩形底色）

	// 系统度量常量 (GetSystemMetrics)
	SM_CXSCREEN = 0 // 屏幕宽度
	SM_CYSCREEN = 1 // 屏幕高度
)

// --- Windows API 结构体映射 ---
// WNDCLASSEX: 窗口类结构体，定义窗口的基本属性
type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr // 窗口过程回调函数指针
	ClsExtra   int32
	WndExtra   int32
	Instance   syscall.Handle // 应用程序实例句柄
	Icon       syscall.Handle
	Cursor     syscall.Handle
	Background syscall.Handle // 背景画刷
	MenuName   *uint16
	ClassName  *uint16 // 窗口类名
	IconSm     syscall.Handle
}

// RECT: 矩形坐标
type RECT struct {
	Left, Top, Right, Bottom int32
}

// PAINTSTRUCT: 绘图信息结构体，用于 BeginPaint/EndPaint
type PAINTSTRUCT struct {
	Hdc         syscall.Handle // 设备上下文句柄
	Erase       int32
	RcPaint     RECT // 需要重绘的矩形区域
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

// MSG: 消息结构体
type MSG struct {
	Hwnd    syscall.Handle
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

// --- DLL 加载与函数指针获取 ---
// 使用 syscall.NewLazyDLL 延迟加载系统 DLL，仅在调用时才解析
var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")

	user32                         = syscall.NewLazyDLL("user32.dll")
	procRegisterClassExW           = user32.NewProc("RegisterClassExW")           // 注册窗口类
	procCreateWindowExW            = user32.NewProc("CreateWindowExW")            // 创建窗口
	procDefWindowProcW             = user32.NewProc("DefWindowProcW")             // 默认窗口过程
	procGetMessageW                = user32.NewProc("GetMessageW")                // 获取消息
	procTranslateMessage           = user32.NewProc("TranslateMessage")           // 翻译消息
	procDispatchMessageW           = user32.NewProc("DispatchMessageW")           // 分发消息
	procBeginPaint                 = user32.NewProc("BeginPaint")                 // 开始绘图
	procEndPaint                   = user32.NewProc("EndPaint")                   // 结束绘图
	procFillRect                   = user32.NewProc("FillRect")                   // 填充矩形
	procInvalidateRect             = user32.NewProc("InvalidateRect")             // 使区域无效（触发重绘）
	procPostQuitMessage            = user32.NewProc("PostQuitMessage")            // 发送退出消息
	procDrawTextW                  = user32.NewProc("DrawTextW")                  // 绘制文本
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")           // 获取屏幕分辨率
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes") // 设置透明属性

	gdi32                = syscall.NewLazyDLL("gdi32.dll")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush") // 创建实心画刷
	procDeleteObject     = gdi32.NewProc("DeleteObject")     // 删除 GDI 对象（防内存泄漏）
	procSetTextColor     = gdi32.NewProc("SetTextColor")     // 设置文本颜色
	procSetBkMode        = gdi32.NewProc("SetBkMode")        // 设置背景混合模式
	procCreateFontW      = gdi32.NewProc("CreateFontW")      // 创建字体
	procSelectObject     = gdi32.NewProc("SelectObject")     // 选择对象到 DC
)

// --- 全局状态变量 ---
var (
	globalHwnd  syscall.Handle      // 全局窗口句柄，用于在 Update 中触发重绘
	displayText string         = "" // 当前显示的文本内容
	mu          sync.Mutex          // 互斥锁，保证多线程修改文本时的安全性
)

// Update 更新显示的文本并触发窗口重绘
// 该函数是线程安全的，可以在其他 Goroutine 中调用
func ScreenUpdateText(text string) {
	mu.Lock()
	displayText = text
	mu.Unlock()
	// 如果窗口已创建，则请求重绘整个客户区
	if globalHwnd != 0 {
		// 第2、3个参数为0表示重绘整个区域，最后一个参数为0表示不擦除背景(由我们在 WM_PAINT 处理)
		procInvalidateRect.Call(uintptr(globalHwnd), 0, 0)
	}
}

func ScreenGetText() string {
	return displayText
}

// Screen 初始化并运行状态窗口
// 注意：该函数包含死循环（消息循环），应在单独的 Goroutine 或主线程末尾运行
func ScreenInit() {
	// Windows GUI 线程必须绑定到特定的 OS 线程，防止 Go 调度器将其切换导致消息丢失
	runtime.LockOSThread()

	// 1. 获取当前模块句柄
	hMod, _, _ := procGetModuleHandleW.Call(0)
	hInstance := syscall.Handle(hMod)

	// 2. 注册窗口类
	className, _ := syscall.UTF16PtrFromString("TransCharClass")

	// 创建 Go 函数到 C 回调的桥接
	wndProcCallback := syscall.NewCallback(wndProc)

	wc := WNDCLASSEX{
		Size:       uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:      CS_HREDRAW | CS_VREDRAW,
		WndProc:    wndProcCallback,
		Instance:   hInstance,
		Background: 0, // 不设置默认背景，在 WM_PAINT 中手动绘制
		ClassName:  className,
	}

	if ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); ret == 0 {
		panic("Failed to register window class")
	}

	// --- 计算窗口位置（屏幕左下角） ---
	// 3.1 获取屏幕高度
	scrHeight, _, _ := procGetSystemMetrics.Call(SM_CYSCREEN)

	// 3.2 定义窗口固定宽高
	// 高度设为 20px，适配 12px 的字体
	winWidth := int32(20)
	winHeight := int32(20)

	// 3.3 计算坐标
	// PosX: 0 (紧贴左侧)
	// PosY: 屏幕高度 - 窗口高度 - 40 (底部留出 40px 边距，避开任务栏上方区域)
	posX := int32(0)
	posY := int32(scrHeight) - winHeight - 40
	// --------------------

	// 4. 创建窗口
	// WS_EX_TOPMOST: 保持在最前
	// WS_EX_TOOLWINDOW: 隐藏任务栏图标
	// WS_EX_LAYERED: 开启透明分层支持
	// WS_POPUP: 无边框
	hwnd, _, _ := procCreateWindowExW.Call(
		WS_EX_TOPMOST|WS_EX_TOOLWINDOW|WS_EX_LAYERED,
		uintptr(unsafe.Pointer(className)),
		0, // 窗口标题（不显示）
		WS_POPUP|WS_VISIBLE,
		uintptr(posX), uintptr(posY), uintptr(winWidth), uintptr(winHeight),
		0, 0, uintptr(hInstance), 0,
	)

	if hwnd == 0 {
		panic("Failed to create window")
	}

	globalHwnd = syscall.Handle(hwnd)

	// 5. 设置透明色键
	// 将颜色 0x000000 (黑色) 设置为完全透明
	// 这意味着在窗口上绘制黑色的地方，实际上会显示出屏幕背景
	procSetLayeredWindowAttributes.Call(hwnd, 0x000000, 0, LWA_COLORKEY)

	// 6. 消息循环
	// 不断从系统消息队列获取消息并分发给 wndProc 处理
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 { // 收到 WM_QUIT
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

// wndProc 窗口过程回调函数，处理系统发送给窗口的消息
func wndProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_PAINT:
		// --- 处理重绘消息 ---
		var ps PAINTSTRUCT
		// 开始绘图，获取设备上下文 (HDC)
		hdc, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps)))

		// 1. 填充背景
		// 创建黑色画刷 (0x000000)
		bgBrush, _, _ := procCreateSolidBrush.Call(0x000000)
		// 用黑色填充背景。由于之前设置了 LWA_COLORKEY 为黑色，这部分区域将变透明
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&ps.RcPaint)), bgBrush)
		procDeleteObject.Call(bgBrush) // 释放 GDI 对象

		// 2. 创建并设置字体
		hFont, _, _ := procCreateFontW.Call(
			12,  // 字体高度
			0,   // 宽度（0=自动匹配）
			0,   // 倾斜角度
			0,   // 基线角度
			400, // 字重 (FW_NORMAL=400, FW_BOLD=700)
			0,   // 斜体
			0,   // 下划线
			0,   // 删除线
			1,   // 字符集 (DEFAULT_CHARSET)
			0,   // 输出精度
			0,   // 剪裁精度
			0,   // 质量
			0,   // 字体族
			0,   // 字体名称指针 (0=使用系统默认)
		)
		// 选择新字体进 DC，并保存旧字体以便恢复
		oldFont, _, _ := procSelectObject.Call(hdc, hFont)

		// 3. 设置文本属性
		procSetTextColor.Call(hdc, 0x00FF00) // 文本颜色：绿色 (G=FF)
		procSetBkMode.Call(hdc, TRANSPARENT) // 文字背景透明（不覆盖文字背后的透明层）

		// 4. 获取当前需要显示的文本
		mu.Lock()
		currentTxt := displayText
		mu.Unlock()

		// 5. 绘制文本
		if len(currentTxt) > 0 {
			utf16Text, _ := syscall.UTF16FromString(currentTxt)
			// DT_CENTER | DT_VCENTER | DT_SINGLELINE: 在矩形内单行居中显示
			procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(&utf16Text[0])), uintptr(len(utf16Text)), uintptr(unsafe.Pointer(&ps.RcPaint)), DT_CENTER|DT_VCENTER|DT_SINGLELINE)
		}

		// 6. 清理 GDI 对象
		procSelectObject.Call(hdc, oldFont)                            // 恢复旧字体
		procDeleteObject.Call(hFont)                                   // 删除临时创建的字体
		procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&ps))) // 结束绘图
		return 0

	case WM_DESTROY:
		// 窗口销毁时，向消息队列发送退出消息，结束消息循环
		procPostQuitMessage.Call(0)
		return 0
	}

	// 对于未处理的消息，交给系统默认处理
	res, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), wParam, lParam)
	return res
}
