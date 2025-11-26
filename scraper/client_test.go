package scraper

import (
	"io"
	"net/http"
	"testing"

	"github.com/fireinrain/javbus-api/config"
)

func TestNewHTTPClient(t *testing.T) {
	var cfg config.Config
	client := NewHTTPClient(&cfg)

	resp, err := client.Get("https://baidu.com")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	defer resp.Body.Close()
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status)
	}
	// 读取并打印响应body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read response body: %v", err)
	}

	// 打印响应内容
	t.Logf("Response body:\n%s", string(body))

}

func TestNewRestyClient(t *testing.T) {
	var cfg config.Config
	client := NewRestyClient(&cfg)
	resp, err := client.R().Get("https://baidu.com")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	// 使用Resty内置方法获取响应体，更简洁
	body := resp.String()

	// 检查响应状态码
	if resp.StatusCode() != http.StatusOK {
		t.Errorf("Expected status OK; got %v", resp.Status())
	}

	// 打印响应内容
	t.Logf("Response body:\n%s", body)
}
