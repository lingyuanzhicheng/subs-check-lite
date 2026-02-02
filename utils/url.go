package utils  
  
import (
    "fmt"
    "io"
    "log/slog"
    "net/http"
    "os"
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

// DownloadRuleYaml 下载规则文件到 output 目录  
func DownloadRuleYaml(outputPath string) error {  
    if config.GlobalConfig.RuleYamlUrl == "" {  
        slog.Info("rule-yaml-url 未配置,跳过规则文件下载")  
        return nil  
    }  
  
    url := WarpUrl(config.GlobalConfig.RuleYamlUrl)  
    slog.Info("开始下载规则文件", "url", url)  
  
    // 创建 HTTP 请求  
    client := &http.Client{  
        Timeout: 30 * time.Second,  
    }  
      
    resp, err := client.Get(url)  
    if err != nil {  
        return fmt.Errorf("下载规则文件失败: %w", err)  
    }  
    defer resp.Body.Close()  
  
    if resp.StatusCode != http.StatusOK {  
        return fmt.Errorf("下载规则文件失败, 状态码: %d", resp.StatusCode)  
    }  
  
    // 读取响应内容  
    body, err := io.ReadAll(resp.Body)  
    if err != nil {  
        return fmt.Errorf("读取规则文件内容失败: %w", err)  
    }  
  
    // 保存到 output 目录  
    ruleFilePath := filepath.Join(outputPath, "rule.yaml")  
    if err := os.WriteFile(ruleFilePath, body, 0644); err != nil {  
        return fmt.Errorf("保存规则文件失败: %w", err)  
    }  
  
    slog.Info("规则文件下载成功", "path", ruleFilePath)  
    return nil  
}

// MergeRuleAndNodeYaml 合并 rule.yaml 和 node.yaml 生成 sub.yaml
func MergeRuleAndNodeYaml(outputPath string) error {
	ruleFilePath := filepath.Join(outputPath, "rule.yaml")
	nodeFilePath := filepath.Join(outputPath, "node.yaml")
	subFilePath := filepath.Join(outputPath, "sub.yaml")

	// 读取 rule.yaml 文件
	ruleData, err := os.ReadFile(ruleFilePath)
	if err != nil {
		return fmt.Errorf("读取 rule.yaml 失败: %w", err)
	}

	// 读取 node.yaml 文件
	nodeData, err := os.ReadFile(nodeFilePath)
	if err != nil {
		return fmt.Errorf("读取 node.yaml 失败: %w", err)
	}

	// 合并内容：rule.yaml 内容在前，node.yaml 内容在后
	mergedContent := string(ruleData) + "\n" + string(nodeData)

	// 写入 sub.yaml 文件
	if err := os.WriteFile(subFilePath, []byte(mergedContent), 0644); err != nil {
		return fmt.Errorf("保存 sub.yaml 失败: %w", err)
	}

	slog.Info("合并规则文件成功", "path", subFilePath)
	return nil
}