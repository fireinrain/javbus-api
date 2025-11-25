package model

import "time"

const (
	JavBusURL = "https://www.javbus.com"

	// JavBusTimeout 爬取页面超时时间
	// 注意：Go 语言中强烈建议直接使用 time.Duration 类型，而不是整数 5000
	// 这样在设置 http.Client.Timeout 时可以直接使用
	JavBusTimeout = 5 * time.Second

	UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"
)
