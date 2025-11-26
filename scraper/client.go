package scraper

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fireinrain/javbus-api/config"
	"github.com/fireinrain/javbus-api/consts"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/proxy"
)

// DefaultHeaderTransport 是一个包装器，用于在请求发出前注入默认 Header
// 这相当于 got.extend({ headers: ... })
type DefaultHeaderTransport struct {
	// 原始的 Transport
	RoundTripper http.RoundTripper
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *DefaultHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 设置 User-Agent (如果请求中未设置)
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", consts.UserAgent)
	}

	// 设置 Accept-Language
	if req.Header.Get("Accept-Language") == "" {
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	}

	// 执行实际请求
	return t.RoundTripper.RoundTrip(req)
}

// NewHTTPClient 创建一个配置好的 HTTP 客户端
func NewHTTPClient(cfg *config.Config) *http.Client {
	// 1. 初始化基础 Transport
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment, // 默认支持系统环境变量
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 2. 解析自定义代理配置 (ENV 覆盖)
	// 逻辑：优先使用 HTTP_PROXY，如果为空则检查 HTTPS_PROXY (模仿原代码逻辑)
	proxyStr := cfg.Proxy.HttpProxy
	if proxyStr == "" {
		proxyStr = cfg.Proxy.HttpsProxy
	}

	if proxyStr != "" {
		u, err := url.Parse(proxyStr)
		if err == nil {
			scheme := strings.ToLower(u.Scheme)

			if strings.HasPrefix(scheme, "socks") {
				// === SOCKS5 代理处理 ===
				// Go 标准库 http.Transport.Proxy 不支持 SOCKS，需要修改 DialContext
				dialer, err := proxy.SOCKS5("tcp", u.Host, nil, proxy.Direct)
				if err == nil {
					baseTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
						// 注意：proxy.SOCKS5 返回的 dialer 不支持 context，但在简单场景下够用
						// 如果需要支持 context 取消，可能需要更复杂的实现，这里保持简单
						return dialer.Dial(network, addr)
					}
				}
			} else if strings.HasPrefix(scheme, "http") {
				// === HTTP/HTTPS 代理处理 ===
				baseTransport.Proxy = http.ProxyURL(u)
			}
		}
	}

	// 3. 组装 Client
	return &http.Client{
		Timeout: consts.JavBusTimeout, // 使用 consts 中定义的 time.Duration
		// 包装 Transport 以注入默认 Header
		Transport: &DefaultHeaderTransport{RoundTripper: baseTransport},
	}
}

//使用resty库

func NewRestyClient(cfg *config.Config) *resty.Client {
	// 创建resty客户端实例
	client := resty.New()

	proxyStr := cfg.Proxy.HttpProxy
	if proxyStr != "" {
		client.SetProxy(proxyStr)
	}
	proxyHttpsStr := cfg.Proxy.HttpsProxy
	if proxyHttpsStr != "" {
		client.SetProxy(proxyHttpsStr)
	}
	proxySocksStr := cfg.Proxy.Socks5Proxy
	if proxyStr != "" {
		client.SetProxy(proxySocksStr)
	}

	// 1. 设置HTTP代理
	//client.SetProxy("http://proxy-server:8080")
	// 2. 设置HTTPS代理
	//client.SetProxy("https://proxy-server:8443")
	// 3. 设置SOCKS5代理
	//client.SetProxy("socks5://proxy-server:1080")
	// 4. 设置带认证的SOCKS5代理
	//client.SetProxy("socks5://username:password@proxy-server:1080")

	// 5. 动态设置代理（根据请求URL决定）
	//client.SetProxyFunction(func(requestURL *url.URL) (*url.URL, error) {
	// 可以根据不同的URL设置不同的代理
	//proxyURL, _ := url.Parse("http://dynamic-proxy:8080")
	//return proxyURL, nil
	//})

	return client
}
