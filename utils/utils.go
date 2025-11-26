package utils

import (
	"strings"

	"github.com/fireinrain/javbus-api/consts"
)

// FormatImageURL 格式化图片链接
// 对应 TS: formatImageUrl(url?: string)
func FormatImageURL(url string) string {
	// 1. 对应 TS 的 url && ... (判空)
	if url == "" {
		return ""
	}

	// 2. 对应 TS 的 !/^http/.test(url)
	// 使用 HasPrefix 替代正则，性能更好
	if !strings.HasPrefix(url, "http") {
		return consts.JavBusURL + url
	}

	return url
}
