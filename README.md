# tkstar 工具包

本项目包含 `tkstar`, `screen`, `hardware`, `network`, `text`, `authorization` 等包，提供了一系列通用工具函数、Windows 屏幕覆盖显示、硬件信息获取、网络请求、加密解密及授权验证功能。

## 目录

- tkstar 包
- screen 包
- hardware 包
- network 包
- text 包
- authorization 包

## tkstar 包

`tkstar` 包提供了一组常用的 Go 语言工具函数，涵盖日志、随机数、文件时间和错误处理等功能。

### 工具函数

#### `func IsDebug() bool`
判断当前运行环境是否为 Debug 模式。
- **原理**：检查可执行文件路径中是否包含 `Temp` 和 `go-build`（通常是 `go run` 生成的临时路径特征）。
- **返回**：`true` (Debug模式) / `false` (Release模式)。

#### `func RandAtomic(max int) int`
生成 `[0, max)` 范围内的随机整数。

#### `func RandAtomicRadius(min, max int) int`
生成 `[min, max]` 范围内的随机整数（包含边界值）。

#### `func CheckErr(err error, errString ...string) bool`
错误检测辅助函数。
- **功能**：如果 `err` 不为 `nil`，打印错误日志并返回 `false`；否则返回 `true`。
- **参数**：`errString` 可选参数，用于在日志中添加自定义错误前缀。

#### `func BuildTime() string`
获取可执行文件的最后修改时间。
- **注意**：这不是编译时的静态时间戳，而是运行时获取的文件最后写入时间。
- **返回**：格式为 `06.0102.1504` 的时间字符串。

#### `func HelperRemoveDuplicates(s []string) []string`
对字符串切片进行去重和去空处理。
- **功能**：移除切片中的空字符串 `""` 和重复的元素，返回新的切片。

#### `func LogFile(logFileName string, systemLog bool) (*logger.Logger, error)`
初始化日志系统。
- **功能**：创建一个写入文件的 Logger 实例。
- **参数**：`logFileName` (日志文件名前缀，不含 .log), `systemLog` (是否同时写入系统事件日志)。
- **返回**：`*logger.Logger` 实例指针和错误信息。

---

## screen 包

`screen` 包用于在 Windows 系统上创建一个透明、置顶、无边框的覆盖窗口，通常用于在屏幕底部显示状态或文本信息。

### 核心函数

#### `func ScreenInit()`
初始化并运行状态窗口。
- **功能**：注册窗口类，创建透明分层窗口（置顶、无任务栏图标），并启动 Windows 消息循环。
- **注意**：该函数包含死循环（消息循环），并且会锁定当前 OS 线程 (`runtime.LockOSThread`)。**必须**在单独的 Goroutine 或主线程末尾运行，否则会阻塞后续代码执行。

#### `func ScreenUpdateText(text string)`
更新屏幕上显示的文本内容并触发重绘。
- **线程安全**：使用了互斥锁，可以在任意 Goroutine 中安全调用。
- **参数**：`text` - 需要显示的字符串内容。

#### `func ScreenGetText() string`
获取当前屏幕上正在显示的文本内容。

---

## hardware 包

`hardware` 包提供硬件信息获取和键盘状态检测功能（Windows）。

### 核心函数

#### `func SysGetSerialKey() string`
获取设备硬件特征码。
- **原理**：结合 MAC 地址、系统 UUID 和硬盘序列号生成唯一的简短机器码。
- **返回**：经过混淆处理的机器码字符串。

#### `func KeyIsPress(keyName string) bool`
判断指定按键是否处于按下状态。
- **参数**：`keyName` - 按键名称（如 "A", "ENTER", "F1", "CTRL" 等，不区分大小写）。
- **原理**：调用 Windows API `GetAsyncKeyState`。

---

## network 包

`network` 包提供网络请求相关功能，支持解析 curl 命令字符串发送请求。

### 核心函数

#### `func NetProxyCurl(proxy, curlBash string) (int, string, error)`
使用指定的代理执行 curl 命令字符串。
- **参数**：`proxy` (代理地址), `curlBash` (curl 命令字符串)。
- **返回**：HTTP状态码, 响应体, 错误信息。

#### `func NetCurl(curlBash string) (int, string, error)`
执行 curl 命令字符串（不使用代理）。
- **功能**：`NetProxyCurl` 的简化版。

#### `func NetParseCurlComd(curlCmd string) (string, string, http.Header, []byte, error)`
解析 curl 命令字符串。
- **功能**：提取请求方法、URL、请求头和请求体。支持多种引号格式。

---

## text 包

`text` 包提供加密解密（RSA, AES）及文本处理功能。

### 核心函数

#### `func TextGetKeyPair(bits int) (*KeyPair, error)`
生成指定位数的 RSA 密钥对。

#### `func TextEncrypt(pub *PublicKey, plaintext []byte) ([]byte, error)`
使用 RSA 公钥对数据进行加密。

#### `func TextDecrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error)`
使用 RSA 私钥对数据进行解密。

#### `func TextVerify(plaintext []byte, decryptedHash []byte) bool`
验证明文的哈希值是否匹配。通常用于签名验证。

#### `func TextAesEncrypt(plainText, key string) (string, error)`
AES 加密。
- **模式**：AES-CFB 模式。
- **输出**：URL 安全的 Base64 编码字符串。

#### `func TextAesDecrypt(cipherText, key string) (string, error)`
AES 解密。
- **输入**：URL 安全的 Base64 编码密文。

---

## authorization 包

`authorization` 包提供简单的远程授权验证功能，通过抓取指定网页表格数据来验证序列号有效期。

### 核心函数

#### `func NewClient(code string) *Client`
创建授权客户端实例。
- **参数**：`code` - 授权页面的标识代码。

#### `func (c *Client) GetAccredit() ([]Accredit, error)`
获取授权信息列表。
- **功能**：访问远程页面，解析 HTML 表格提取序列号和过期时间。包含重试机制。

#### `func (c *Client) CheckAccredit(key string) bool`
检查指定序列号是否有效且未过期。
- **功能**：优先使用缓存数据，如果缓存为空则请求网络。验证序列号是否存在以及当前时间是否在有效期内。
