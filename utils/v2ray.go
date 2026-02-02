package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/beck-8/subs-check/config"
	"gopkg.in/yaml.v3"
)

// ConvertToV2Ray 将 node.yaml 转换为 V2Ray 订阅格式
func ConvertToV2Ray(outputPath string) error {
	// 检查配置开关
	if !config.GlobalConfig.V2RaySubscription {
		slog.Debug("V2Ray 订阅转换已禁用 跳过")
		return nil
	}

	allYamlPath := filepath.Join(outputPath, "node.yaml")
	yamlData, err := os.ReadFile(allYamlPath)
	if err != nil {
		return fmt.Errorf("读取 node.yaml 失败: %w", err)
	}

	var yamlConfig struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(yamlData, &yamlConfig); err != nil {
		return fmt.Errorf("解析 YAML 失败: %w", err)
	}

	var v2rayLinks []string
	for _, proxy := range yamlConfig.Proxies {
		proxyType, _ := proxy["type"].(string)

		var link string
		switch proxyType {
		case "vmess":
			link = convertVMessToLink(proxy)
		case "vless":
			link = convertVLESSToLink(proxy)
		case "ss", "shadowsocks":
			link = convertShadowsocksToLink(proxy)
		case "trojan":
			link = convertTrojanToLink(proxy)
		case "hysteria2", "hy2":
			link = convertHysteria2ToLink(proxy)
		}

		if link != "" {
			v2rayLinks = append(v2rayLinks, link)
		}
	}

	if len(v2rayLinks) == 0 {
		slog.Warn("没有可转换的 V2Ray 节点")
		return nil
	}

	// 直接保存链接列表,不进行 Base64 编码
	v2rayContent := strings.Join(v2rayLinks, "\n")

	v2rayPath := filepath.Join(outputPath, "v2ray.txt")
	if err := os.WriteFile(v2rayPath, []byte(v2rayContent), 0644); err != nil {
		return fmt.Errorf("保存 V2Ray 订阅失败: %w", err)
	}

	slog.Info("V2Ray 订阅转换成功", "path", v2rayPath, "节点数", len(v2rayLinks))
	return nil
}

// removeFlagEmoji 移除 Unicode 旗帜字符(仅移除开头的旗帜)
func removeFlagEmoji(name string) string {
	// 移除开头的区域指示符号 (U+1F1E6 到 U+1F1FF)
	re := regexp.MustCompile(`^[\x{1F1E6}-\x{1F1FF}]+`)
	return strings.TrimSpace(re.ReplaceAllString(name, ""))
}

// convertVMessToLink 转换 VMess 为链接格式
func convertVMessToLink(proxy map[string]any) string {
	name := removeFlagEmoji(getString(proxy, "name"))
	server := getString(proxy, "server")
	port := proxy["port"]
	uuid := getString(proxy, "uuid")

	if server == "" || uuid == "" || port == nil {
		return ""
	}

	vmessConfig := map[string]any{
		"v":    "2",
		"ps":   name,
		"add":  server,
		"port": fmt.Sprintf("%v", port),
		"id":   uuid,
		"aid":  fmt.Sprintf("%v", getInt(proxy, "alterId")),
		"scy":  getString(proxy, "cipher"),
		"net":  getString(proxy, "network"),
		"type": getString(proxy, "type"),
		"host": getNestedString(proxy, "ws-opts", "headers", "Host"),
		"path": getNestedString(proxy, "ws-opts", "path"),
		"tls":  getTLS(proxy),
		"sni":  getString(proxy, "servername", "sni"),
		"alpn": getALPN(proxy),
		"fp":   getString(proxy, "client-fingerprint"),
	}

	// 移除空值
	for k, v := range vmessConfig {
		if v == "" || v == "0" {
			delete(vmessConfig, k)
		}
	}

	jsonBytes, _ := json.Marshal(vmessConfig)
	encoded := base64.StdEncoding.EncodeToString(jsonBytes)
	return "vmess://" + encoded
}

// convertVLESSToLink 转换 VLESS 为链接格式
func convertVLESSToLink(proxy map[string]any) string {
	name := removeFlagEmoji(getString(proxy, "name"))
	server := getString(proxy, "server")
	port := fmt.Sprintf("%v", proxy["port"])
	uuid := getString(proxy, "uuid")

	if server == "" || uuid == "" || port == "" {
		return ""
	}

	params := url.Values{}
	params.Set("encryption", "none")

	if flow := getString(proxy, "flow"); flow != "" {
		params.Set("flow", flow)
	}

	if network := getString(proxy, "network"); network != "" {
		params.Set("type", network)
	}

	if security := getString(proxy, "tls"); security != "" {
		params.Set("security", security)
	}

	if sni := getString(proxy, "servername", "sni"); sni != "" {
		params.Set("sni", sni)
	}

	if fp := getString(proxy, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	if alpn := getALPN(proxy); alpn != "" {
		params.Set("alpn", alpn)
	}

	if skipVerify := getBool(proxy, "skip-cert-verify"); skipVerify {
		params.Set("allowInsecure", "1")
	}

	if network := getString(proxy, "network"); network == "ws" {
		if host := getNestedString(proxy, "ws-opts", "headers", "Host"); host != "" {
			params.Set("host", host)
		}
		if path := getNestedString(proxy, "ws-opts", "path"); path != "" {
			params.Set("path", path)
		}
	}

	link := fmt.Sprintf("vless://%s@%s:%s?%s#%s",
		uuid, server, port, params.Encode(), url.QueryEscape(name))
	return link
}

// convertTrojanToLink 转换 Trojan 为链接格式
func convertTrojanToLink(proxy map[string]any) string {
	name := removeFlagEmoji(getString(proxy, "name"))
	server := getString(proxy, "server")
	port := fmt.Sprintf("%v", proxy["port"])
	password := getString(proxy, "password")

	if server == "" || password == "" || port == "" {
		return ""
	}

	params := url.Values{}
	params.Set("security", "tls")

	if sni := getString(proxy, "sni", "servername"); sni != "" {
		params.Set("sni", sni)
	}

	if fp := getString(proxy, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	if alpn := getALPN(proxy); alpn != "" {
		params.Set("alpn", alpn)
	}

	if network := getString(proxy, "network"); network != "" {
		params.Set("type", network)
	} else {
		params.Set("type", "tcp")
	}

	if skipVerify := getBool(proxy, "skip-cert-verify"); skipVerify {
		params.Set("allowInsecure", "1")
	}

	link := fmt.Sprintf("trojan://%s@%s:%s?%s#%s",
		password, server, port, params.Encode(), url.QueryEscape(name))
	return link
}

// convertHysteria2ToLink 转换 Hysteria2 为链接格式
func convertHysteria2ToLink(proxy map[string]any) string {
	name := removeFlagEmoji(getString(proxy, "name"))
	server := getString(proxy, "server")
	port := fmt.Sprintf("%v", proxy["port"])
	password := getString(proxy, "password")

	if server == "" || password == "" || port == "" {
		return ""
	}

	params := url.Values{}

	if sni := getString(proxy, "sni", "servername"); sni != "" {
		params.Set("sni", sni)
	}

	if obfs := getString(proxy, "obfs"); obfs != "" {
		params.Set("obfs", obfs)
		if obfsPassword := getString(proxy, "obfs-password"); obfsPassword != "" {
			params.Set("obfs-password", obfsPassword)
		}
	}

	if skipVerify := getBool(proxy, "skip-cert-verify"); skipVerify {
		params.Set("insecure", "1")
	}

	if fp := getString(proxy, "client-fingerprint"); fp != "" {
		params.Set("fp", fp)
	}

	if alpn := getALPN(proxy); alpn != "" {
		params.Set("alpn", alpn)
	}

	link := fmt.Sprintf("hysteria2://%s@%s:%s?%s#%s",
		password, server, port, params.Encode(), url.QueryEscape(name))
	return link
}

// convertShadowsocksToLink 转换 Shadowsocks 为链接格式
func convertShadowsocksToLink(proxy map[string]any) string {
	name := removeFlagEmoji(getString(proxy, "name"))
	server := getString(proxy, "server")
	port := fmt.Sprintf("%v", proxy["port"])
	password := getString(proxy, "password")
	cipher := getString(proxy, "cipher")

	if server == "" || password == "" || port == "" {
		return ""
	}

	if cipher == "" {
		cipher = "aes-256-gcm"
	}

	userInfo := fmt.Sprintf("%s:%s", cipher, password)
	encoded := base64.StdEncoding.EncodeToString([]byte(userInfo))

	link := fmt.Sprintf("ss://%s@%s:%s#%s",
		encoded, server, port, url.QueryEscape(name))
	return link
}

// 辅助函数

// getString 从 map 中获取字符串值,支持多个备选键
func getString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}

// getNestedString 获取嵌套 map 中的字符串值
func getNestedString(m map[string]any, keys ...string) string {
	current := m
	for i, key := range keys {
		if i == len(keys)-1 {
			if val, ok := current[key]; ok {
				if str, ok := val.(string); ok {
					return str
				}
			}
			return ""
		}
		if val, ok := current[key]; ok {
			if nested, ok := val.(map[string]any); ok {
				current = nested
			} else {
				return ""
			}
		} else {
			return ""
		}
	}
	return ""
}

// getInt 从 map 中获取整数值
func getInt(m map[string]any, key string) int {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return 0
}

// getBool 从 map 中获取布尔值
func getBool(m map[string]any, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// getTLS 获取 TLS 配置
func getTLS(m map[string]any) string {
	if tls := getString(m, "tls"); tls == "true" || tls == "tls" {
		return "tls"
	}
	return ""
}

// getALPN 获取 ALPN 配置
func getALPN(m map[string]any) string {
	if alpnList, ok := m["alpn"].([]any); ok && len(alpnList) > 0 {
		alpns := make([]string, 0, len(alpnList))
		for _, a := range alpnList {
			if s, ok := a.(string); ok {
				alpns = append(alpns, s)
			}
		}
		return strings.Join(alpns, ",")
	}
	return ""
}