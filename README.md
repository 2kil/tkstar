# tkstar

`tkstar` 是一个 Go 工具包集合，包含以下能力：

- `tkstar`：通用辅助函数
- `authorization`：远程授权校验
- `network`：从 curl 字符串发起请求
- `text`：RSA/AES 与文本处理
- `screen`：Windows 屏幕悬浮文本
- `hardware`：Windows 硬件和按键检测
- `edge`：基于 chromedp 的 Edge 浏览器控制

## 安装

```bash
go get github.com/2Kil/tkstar
```

## 包说明

- `screen`、`hardware`、`edge` 依赖 Windows API，仅适合 Windows 环境。
- `screen.ScreenInit()` 会进入消息循环，必须放在 goroutine 中或主线程最后执行。
- `edge` 包依赖本机安装 Microsoft Edge。
- `authorization` 依赖远程二维码页面格式，示例中的地址和密码请替换为实际值。

## tkstar 包

导入：

```go
import "github.com/2Kil/tkstar"
```

### `func IsDebug() bool`

判断当前是否接近 `go run` 的临时构建环境。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	fmt.Println("debug:", tkstar.IsDebug())
}
```

### `func RandAtomic(max int) int`

返回 `[0, max)` 的随机整数。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	n := tkstar.RandAtomic(10)
	fmt.Println(n)
}
```

### `func RandAtomicRadius(min, max int) int`

返回 `[min, max]` 的随机整数。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	n := tkstar.RandAtomicRadius(5, 12)
	fmt.Println(n)
}
```

### `func CheckErr(err error, errString ...string) bool`

统一打印错误并返回是否成功。

```go
package main

import (
	"errors"
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	ok := tkstar.CheckErr(errors.New("open failed"), "load config")
	fmt.Println("ok:", ok)
}
```

### `func BuildTime() string`

返回当前可执行文件最后修改时间。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	fmt.Println("build time:", tkstar.BuildTime())
}
```

### `func HelperRemoveDuplicates(s []string) []string`

去重并移除空字符串。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar"
)

func main() {
	items := []string{"A", "", "B", "A", "C", "B"}
	fmt.Println(tkstar.HelperRemoveDuplicates(items))
}
```

### `func LogFile(logFileName string, systemLog bool) (*logger.Logger, error)`

初始化日志文件。

```go
package main

import (
	"log"

	"github.com/2Kil/tkstar"
)

func main() {
	l, err := tkstar.LogFile("app", false)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	l.Info("service started")
}
```

## authorization 包

导入：

```go
import "github.com/2Kil/tkstar/authorization"
```

### `func NewClient(code string, pwd ...string) *Client`

创建授权客户端。新版活码可同时传密码。

```go
package main

import "github.com/2Kil/tkstar/authorization"

func main() {
	_ = authorization.NewClient("active.clewm.net/q8tDtnl")
	_ = authorization.NewClient("qr61.cn/o78kxB/q8tDtnl", "123456")
}
```

### `func (c *Client) GetAccredit() ([]Accredit, error)`

按网页解析方式获取授权表格。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/authorization"
)

func main() {
	client := authorization.NewClient("active.clewm.net/q8tDtnl")
	list, err := client.GetAccredit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list)
}
```

### `func (c *Client) GetAccredit2() ([]Accredit, error)`

按接口方式获取新版活码授权数据。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/authorization"
)

func main() {
	client := authorization.NewClient("qr61.cn/o78kxB/q8tDtnl", "123456")
	list, err := client.GetAccredit2()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list)
}
```

### `func (c *Client) CheckAccredit(key string) bool`

检查某个序列号是否未过期。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar/authorization"
)

func main() {
	client := authorization.NewClient("qr61.cn/o78kxB/q8tDtnl", "123456")
	fmt.Println(client.CheckAccredit("DEVICE-001"))
}
```

## network 包

导入：

```go
import "github.com/2Kil/tkstar/network"
```

### `func NetProxyCurl(proxy, curlBash string) (int, string, error)`

使用代理执行 curl 字符串。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/network"
)

func main() {
	code, body, err := network.NetProxyCurl(
		"http://127.0.0.1:7890",
		`curl "https://httpbin.org/get" -H "Accept: application/json"`,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
	fmt.Println(body)
}
```

### `func NetParseCurlComd(curlCmd string) (string, string, http.Header, []byte, error)`

只解析，不发送请求。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/network"
)

func main() {
	method, urlStr, headers, body, err := network.NetParseCurlComd(
		`curl "https://example.com/api" -X POST -H "Content-Type: application/json" --data-raw "{\"name\":\"tkstar\"}"`,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(method)
	fmt.Println(urlStr)
	fmt.Println(headers.Get("Content-Type"))
	fmt.Println(string(body))
}
```

### `func NetCurl(curlBash string) (int, string, error)`

不使用代理执行 curl 字符串。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/network"
)

func main() {
	code, body, err := network.NetCurl(`curl "https://httpbin.org/get"`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
	fmt.Println(body)
}
```

## text 包

导入：

```go
import "github.com/2Kil/tkstar/text"
```

### `func TextGetKeyPair(bits int) (*KeyPair, error)`

生成 RSA 密钥对。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/text"
)

func main() {
	keyPair, err := text.TextGetKeyPair(1024)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(keyPair.PublicKey.N)
}
```

### `func TextEncrypt(pub *PublicKey, plaintext []byte) ([]byte, error)`

使用公钥加密。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/text"
)

func main() {
	keyPair, err := text.TextGetKeyPair(1024)
	if err != nil {
		log.Fatal(err)
	}

	cipherText, err := text.TextEncrypt(keyPair.PublicKey, []byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cipherText)
}
```

### `func TextDecrypt(priv *PrivateKey, ciphertext []byte) ([]byte, error)`

使用私钥解密。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/text"
)

func main() {
	keyPair, err := text.TextGetKeyPair(1024)
	if err != nil {
		log.Fatal(err)
	}

	cipherText, err := text.TextEncrypt(keyPair.PublicKey, []byte("hello"))
	if err != nil {
		log.Fatal(err)
	}

	plainText, err := text.TextDecrypt(keyPair.PrivateKey, cipherText)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(plainText))
}
```

### `func TextVerify(plaintext []byte, decryptedHash []byte) bool`

比较明文 SHA-256 和“解密后的哈希值”是否一致。

```go
package main

import (
	"crypto/sha256"
	"fmt"

	"github.com/2Kil/tkstar/text"
)

func main() {
	plain := []byte("hello")
	hash := sha256.Sum256(plain)
	fmt.Println(text.TextVerify(plain, hash[:]))
}
```

### `func TextAesEncrypt(plainText, key string) (string, error)`

使用 AES-CFB 加密。密钥长度必须是 16、24 或 32 字节。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/text"
)

func main() {
	cipherText, err := text.TextAesEncrypt("hello", "1234567890abcdef")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cipherText)
}
```

### `func TextAesDecrypt(cipherText, key string) (string, error)`

解密 `TextAesEncrypt` 返回的数据。

```go
package main

import (
	"fmt"
	"log"

	"github.com/2Kil/tkstar/text"
)

func main() {
	cipherText, err := text.TextAesEncrypt("hello", "1234567890abcdef")
	if err != nil {
		log.Fatal(err)
	}

	plainText, err := text.TextAesDecrypt(cipherText, "1234567890abcdef")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(plainText)
}
```

## screen 包

导入：

```go
import "github.com/2Kil/tkstar/screen"
```

### `func ScreenInit()`

初始化透明悬浮窗。

```go
package main

import "github.com/2Kil/tkstar/screen"

func main() {
	go screen.ScreenInit()
	select {}
}
```

### `func ScreenUpdateText(text string)`

更新悬浮窗文本。

```go
package main

import (
	"time"

	"github.com/2Kil/tkstar/screen"
)

func main() {
	go screen.ScreenInit()
	time.Sleep(time.Second)
	screen.ScreenUpdateText("RUN")
	select {}
}
```

### `func ScreenGetText() string`

读取当前显示文本。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar/screen"
)

func main() {
	screen.ScreenUpdateText("READY")
	fmt.Println(screen.ScreenGetText())
}
```

## hardware 包

导入：

```go
import "github.com/2Kil/tkstar/hardware"
```

### `func SysGetSerialKey() string`

生成设备特征码。

```go
package main

import (
	"fmt"

	"github.com/2Kil/tkstar/hardware"
)

func main() {
	fmt.Println(hardware.SysGetSerialKey())
}
```

### `func KeyIsPress(keyName string) bool`

检测某个键当前是否按下。

```go
package main

import (
	"fmt"
	"time"

	"github.com/2Kil/tkstar/hardware"
)

func main() {
	for {
		if hardware.KeyIsPress("F8") {
			fmt.Println("F8 pressed")
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}
```

## edge 包

导入：

```go
import tkEdge "github.com/2Kil/tkstar/edge"
```

### `func Run(urlPath string, msedgePath ...string)`

启动带界面的 Edge。

```go
package main

import tkEdge "github.com/2Kil/tkstar/edge"

func main() {
	tkEdge.Run("https://example.com")
	select {}
}
```

### `func RunCli(urlPath string, msedgePath ...string)`

启动无头 Edge。

```go
package main

import tkEdge "github.com/2Kil/tkstar/edge"

func main() {
	tkEdge.RunCli("https://example.com")
	select {}
}
```

### `func Stop()`

停止当前浏览器上下文。

```go
package main

import (
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://example.com")
	time.Sleep(5 * time.Second)
	tkEdge.Stop()
}
```

### `func LoadUrl(targetURL string)`

让已启动的浏览器跳转到新地址。

```go
package main

import (
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://example.com")
	time.Sleep(3 * time.Second)
	tkEdge.LoadUrl("https://httpbin.org/get")
	select {}
}
```

### `func GetUrl() (string, error)`

读取当前页面地址。

```go
package main

import (
	"fmt"
	"log"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://example.com")
	time.Sleep(3 * time.Second)

	urlStr, err := tkEdge.GetUrl()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(urlStr)
}
```

### `func GetCookies() (map[string]string, error)`

获取当前页面 Cookie。

```go
package main

import (
	"fmt"
	"log"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://example.com")
	time.Sleep(3 * time.Second)

	cookies, err := tkEdge.GetCookies()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cookies)
}
```

### `func GetCookiesAll() (map[string]string, error)`

从 storage 读取全部 Cookie。

```go
package main

import (
	"fmt"
	"log"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://example.com")
	time.Sleep(3 * time.Second)

	cookies, err := tkEdge.GetCookiesAll()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cookies)
}
```

### `func GetReq(key string) string`

获取最近一次匹配请求中的请求头。

```go
package main

import (
	"fmt"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://httpbin.org/headers")
	time.Sleep(5 * time.Second)
	fmt.Println(tkEdge.GetReq("User-Agent"))
}
```

### `func GetRes(key string) string`

获取最近一次匹配响应中的响应头。

```go
package main

import (
	"fmt"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://httpbin.org/get")
	time.Sleep(5 * time.Second)
	fmt.Println(tkEdge.GetRes("content-type"))
}
```

### `func GetUrlQuery(key string) string`

获取最近一次匹配请求 URL 里的查询参数。

```go
package main

import (
	"fmt"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://httpbin.org/get?token=123")
	time.Sleep(5 * time.Second)
	fmt.Println(tkEdge.GetUrlQuery("token"))
}
```

### `func Clear()`

清理最近一次保存的请求/响应信息。

```go
package main

import (
	"fmt"
	"time"

	tkEdge "github.com/2Kil/tkstar/edge"
)

func main() {
	tkEdge.Run("https://httpbin.org/get?token=123")
	time.Sleep(5 * time.Second)

	fmt.Println(tkEdge.GetUrlQuery("token"))
	tkEdge.Clear()
	fmt.Println(tkEdge.GetUrlQuery("token"))
}
```

## 许可证

本项目采用 [LICENSE](./LICENSE) 中定义的许可证。
