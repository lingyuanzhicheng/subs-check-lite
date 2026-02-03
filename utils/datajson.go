package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/beck-8/subs-check/config"
	"gopkg.in/yaml.v3"
)

// StatsData 统计数据结构
type StatsData struct {
	TotalNodes        int            `json:"nodes"`
	Countries         map[string]int `json:"countries"`
	Types             map[string]int `json:"types"`
	V2RaySubscription bool           `json:"v2ray-subscription"`
	MediaCheck        bool           `json:"media-check"`
}

// ProxiesYAML 用于解析 node.yaml 的结构
type ProxiesYAML struct {
	Proxies []map[string]interface{} `json:"proxies" yaml:"proxies"`
}

// GenerateStatsJSON 生成统计数据 JSON 文件
func GenerateStatsJSON(outputPath string) error {
	yamlPath := filepath.Join(outputPath, "node.yaml")
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("读取 node.yaml 失败: %w", err)
	}

	var proxiesData ProxiesYAML
	if err := yaml.Unmarshal(yamlData, &proxiesData); err != nil {
		return fmt.Errorf("解析 YAML 失败: %w", err)
	}

	stats := StatsData{
		Countries: make(map[string]int),
		Types:     make(map[string]int),
	}

	countryRegex := regexp.MustCompile(`^([\x{1F1E6}-\x{1F1FF}]{2})([A-Z]{2})`)

	for _, proxy := range proxiesData.Proxies {
		stats.TotalNodes++

		if name, ok := proxy["name"].(string); ok {
			matches := countryRegex.FindStringSubmatch(name)
			if len(matches) > 2 {
				countryCode := matches[2]
				stats.Countries[countryCode]++
			}
		}

		if proxyType, ok := proxy["type"].(string); ok {
			stats.Types[proxyType]++
		}
	}

	stats.V2RaySubscription = config.GlobalConfig.V2RaySubscription
	stats.MediaCheck = config.GlobalConfig.MediaCheck

	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	jsonPath := filepath.Join(outputPath, "stats.json")
	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("保存 stats.json 失败: %w", err)
	}

	return nil
}
