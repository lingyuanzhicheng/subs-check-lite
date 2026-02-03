package utils

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/beck-8/subs-check/config"
)

// WarpUrl 处理订阅 URL,支持时间占位符和 GitHub 代理
func WarpUrl(url string) string {
	url = formatTimePlaceholders(url, time.Now())

	// 如果url中以https://raw.githubusercontent.com开头，那么就使用github代理
	if strings.HasPrefix(url, "https://raw.githubusercontent.com") {
		return config.GlobalConfig.GithubProxy + url
	}
	return url
}

// formatTimePlaceholders 动态时间占位符
// 支持在链接中使用时间占位符，会自动替换成当前日期/时间:
// - `{Y}` - 四位年份 (2023)
// - `{m}` - 两位月份 (01-12)
// - `{d}` - 两位日期 (01-31)
// - `{Ymd}` - 组合日期 (20230131)
// - `{Y_m_d}` - 下划线分隔 (2023_01_31)
// - `{Y-m-d}` - 横线分隔 (2023-01-31)
func formatTimePlaceholders(url string, t time.Time) string {
	replacer := strings.NewReplacer(
		"{Y}", t.Format("2006"),
		"{m}", t.Format("01"),
		"{d}", t.Format("02"),
		"{Ymd}", t.Format("20060102"),
		"{Y_m_d}", t.Format("2006_01_02"),
		"{Y-m-d}", t.Format("2006-01-02"),
	)
	return replacer.Replace(url)
}

// GetConfigDir 获取配置文件所在目录
func GetConfigDir() string {
	basePath := GetExecutablePath()
	configDir := filepath.Join(basePath, "config")
	return configDir
}
