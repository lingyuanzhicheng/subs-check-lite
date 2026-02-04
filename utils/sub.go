package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/beck-8/subs-check/config"
	"gopkg.in/yaml.v3"
)

// CountryInfo 国家信息结构
type CountryInfo struct {
	Code   string `json:"code"`
	Flag   string `json:"flag"`
	CnName string `json:"cn-name"`
	EnName string `json:"en-name"`
}

// ConfigData 配置数据结构
type ConfigData struct {
	MediaCheck bool     `yaml:"media-check"`
	Platforms  []string `yaml:"platforms"`
}

// GenerateSubYAML 生成 sub.yaml 文件
func GenerateSubYAML(outputPath string) error {
	// 读取 rule.yaml
	rulePath := filepath.Join(outputPath, "..", "config", "rule.yaml")
	ruleContent, err := os.ReadFile(rulePath)
	if err != nil {
		return fmt.Errorf("读取 rule.yaml 失败: %w", err)
	}

	// 读取 stats.json
	statsPath := filepath.Join(outputPath, "stats.json")
	statsData, err := readStatsData(statsPath)
	if err != nil {
		return fmt.Errorf("读取 stats.json 失败: %w", err)
	}

	// 读取 countries.json
	countriesPath := filepath.Join(outputPath, "..", "config", "countries.json")
	countriesMap, err := readCountriesMap(countriesPath)
	if err != nil {
		return fmt.Errorf("读取 countries.json 失败: %w", err)
	}

	// 读取 config.yaml
	configPath := filepath.Join(outputPath, "..", "config", "config.yaml")
	configData, err := readConfigData(configPath)
	if err != nil {
		return fmt.Errorf("读取 config.yaml 失败: %w", err)
	}

	// 读取 node.yaml
	nodePath := filepath.Join(outputPath, "node.yaml")
	nodeContent, err := os.ReadFile(nodePath)
	if err != nil {
		return fmt.Errorf("读取 node.yaml 失败: %w", err)
	}

	// 处理生成 sub.yaml
	subContent, err := processRuleContent(string(ruleContent), statsData, countriesMap, configData, string(nodeContent))
	if err != nil {
		return fmt.Errorf("处理 rule.yaml 失败: %w", err)
	}

	// 写入 sub.yaml
	subPath := filepath.Join(outputPath, "sub.yaml")
	if err := os.WriteFile(subPath, []byte(subContent), 0644); err != nil {
		return fmt.Errorf("保存 sub.yaml 失败: %w", err)
	}

	return nil
}

// readStatsData 读取统计数据
func readStatsData(path string) (*StatsData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var stats StatsData
	if err := yaml.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// readCountriesMap 读取国家映射
func readCountriesMap(path string) (map[string]CountryInfo, error) {
	var data []byte
	var err error

	data, err = os.ReadFile(path)
	if err != nil {
		if len(config.DefaultCountriesTemplate) > 0 {
			data = config.DefaultCountriesTemplate
		} else {
			return nil, err
		}
	}

	var countries []CountryInfo
	if err := json.Unmarshal(data, &countries); err != nil {
		return nil, err
	}

	countriesMap := make(map[string]CountryInfo)
	for _, country := range countries {
		countriesMap[country.Code] = country
	}

	return countriesMap, nil
}

// readConfigData 读取配置数据
func readConfigData(path string) (*ConfigData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ConfigData
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// processRuleContent 处理 rule.yaml 内容
func processRuleContent(ruleContent string, statsData *StatsData, countriesMap map[string]CountryInfo, configData *ConfigData, nodeContent string) (string, error) {
	lines := strings.Split(ruleContent, "\n")
	var result []string
	i := 0

	for i < len(lines) {
		line := lines[i]

		// 处理 {countries.name.list}
		if strings.Contains(line, "{countries.name.list}") {
			indent := getIndent(line)
			countryList := generateCountryNameList(statsData.Countries, countriesMap, indent)
			result = append(result, countryList...)
			i++
			continue
		}

		// 处理 {countries.list}
		if strings.Contains(line, "{countries.list}") {
			indent := getIndent(line)
			countryGroups := generateCountryGroups(statsData.Countries, countriesMap, indent)
			result = append(result, countryGroups...)
			i++
			continue
		}

		// 处理 {media.list}
		if strings.Contains(line, "{media.list}") {
			indent := getIndent(line)
			mediaGroups := generateMediaGroups(configData, indent)
			if len(mediaGroups) > 0 {
				result = append(result, mediaGroups...)
			}
			i++
			continue
		}

		result = append(result, line)
		i++
	}

	// 添加 node.yaml 内容
	result = append(result, "")
	result = append(result, strings.Split(nodeContent, "\n")...)

	return strings.Join(result, "\n"), nil
}

// getIndent 获取行缩进
func getIndent(line string) string {
	indent := ""
	for _, c := range line {
		if c == ' ' || c == '\t' {
			indent += string(c)
		} else {
			break
		}
	}
	return indent
}

// generateCountryNameList 生成国家名称列表
func generateCountryNameList(countries map[string]int, countriesMap map[string]CountryInfo, indent string) []string {
	var result []string

	for code := range countries {
		if country, ok := countriesMap[code]; ok {
			result = append(result, fmt.Sprintf("%s- %s %s", indent, country.Flag, country.CnName))
		}
	}

	return result
}

// generateCountryGroups 生成国家代理组
func generateCountryGroups(countries map[string]int, countriesMap map[string]CountryInfo, indent string) []string {
	var result []string

	for code := range countries {
		if country, ok := countriesMap[code]; ok {
			group := fmt.Sprintf(
				"%s- name: %s %s\n%s  include-all: true\n%s  filter: (?i)%s|%s|%s_|%s\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50",
				indent, country.Flag, country.CnName,
				indent,
				indent, country.Flag, country.CnName, code, country.EnName,
				indent,
				indent,
				indent)
			result = append(result, strings.Split(group, "\n")...)
		}
	}

	return result
}

// generateMediaGroups 生成媒体代理组
func generateMediaGroups(configData *ConfigData, indent string) []string {
	var result []string

	if !configData.MediaCheck {
		return result
	}

	mediaTemplates := map[string]string{
		"tiktok":  `TikTok\n%s  include-all: true\n%s  filter: (?i)抖音|TK-|TikTok\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
		"youtube": `YouTube\n%s  include-all: true\n%s  filter: (?i)油管|YT-|YouTube\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
		"netflix": `Netflix\n%s  include-all: true\n%s  filter: (?i)奈菲|NF|Netflix\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
		"disney":  `Disney\n%s  include-all: true\n%s  filter: (?i)迪士尼|D+|Disney\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
		"openai":  `OpenAI\n%s  include-all: true\n%s  filter: (?i)GPT|OpenAI\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
		"gemini":  `Gemini\n%s  include-all: true\n%s  filter: (?i)GM|Gemini\n%s  type: url-test\n%s  interval: 300\n%s  tolerance: 50`,
	}

	for _, platform := range configData.Platforms {
		if platform == "iprisk" {
			continue
		}
		if template, ok := mediaTemplates[platform]; ok {
			group := fmt.Sprintf("%s- name: %s", indent, fmt.Sprintf(template, indent, indent, indent, indent, indent))
			result = append(result, strings.Split(group, "\n")...)
		}
	}

	return result
}
