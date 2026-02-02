package save

import (
	"fmt"
	"log/slog"

	"github.com/beck-8/subs-check/check"
	"github.com/beck-8/subs-check/config"
	"github.com/beck-8/subs-check/save/method"
	"github.com/beck-8/subs-check/utils"
	"gopkg.in/yaml.v3"
)

// ProxyCategory 定义代理分类
type ProxyCategory struct {
	Name    string
	Proxies []map[string]any
	Filter  func(result check.Result) bool
}

// ConfigSaver 处理配置保存的结构体
type ConfigSaver struct {
	results    []check.Result
	categories []ProxyCategory
	saveMethod func([]byte, string) error
}

// NewConfigSaver 创建新的配置保存器
func NewConfigSaver(results []check.Result) *ConfigSaver {
	return &ConfigSaver{
		results:    results,
		saveMethod: chooseSaveMethod(),
		categories: []ProxyCategory{
			{
				Name:    "node.yaml",
				Proxies: make([]map[string]any, 0),
				Filter:  func(result check.Result) bool { return true },
			},
		},
	}
}

// SaveConfig 保存配置的入口函数
func SaveConfig(results []check.Result) {
	tmp := config.GlobalConfig.SaveMethod
	config.GlobalConfig.SaveMethod = "local"
	// 奇技淫巧，保存到本地一份，因为我没想道其他更好的方法同时保存
	{
		saver := NewConfigSaver(results)
		if err := saver.Save(); err != nil {
			slog.Error(fmt.Sprintf("保存配置失败: %v", err))
		}
	}

	if tmp == "local" {
		return
	}
	config.GlobalConfig.SaveMethod = tmp
	// 如果其他配置验证失败，还会保存到本地一次
	{
		saver := NewConfigSaver(results)
		if err := saver.Save(); err != nil {
			slog.Error(fmt.Sprintf("保存配置失败: %v", err))
		}
	}
}

// Save 执行保存操作
func (cs *ConfigSaver) Save() error {
	// 分类处理代理
	cs.categorizeProxies()

	// 保存各个类别的代理
	for _, category := range cs.categories {
		if err := cs.saveCategory(category); err != nil {
			slog.Error(fmt.Sprintf("保存到%s失败: %v", config.GlobalConfig.SaveMethod, err))
			continue
		}
	}

	// 下载规则文件 (新增)
	saver, err := method.NewLocalSaver()

	if err != nil {
		slog.Error(fmt.Sprintf("创建本地保存器失败: %v", err))
	} else {
		if err := utils.DownloadRuleYaml(saver.OutputPath); err != nil {
			slog.Error(fmt.Sprintf("下载规则文件失败: %v", err))
		}

		// 合并 rule.yaml 和 node.yaml 生成 sub.yaml
		if err := utils.MergeRuleAndNodeYaml(saver.OutputPath); err != nil {
			slog.Error(fmt.Sprintf("合并规则文件失败: %v", err))
		}

		// 生成转换订阅 V2ray
		if err := utils.ConvertToV2Ray(saver.OutputPath); err != nil {
			slog.Error(fmt.Sprintf("转换 V2Ray 订阅失败: %v", err))
		}

		// 生成统计数据 JSON
		if err := utils.GenerateStatsJSON(saver.OutputPath); err != nil {
			slog.Error(fmt.Sprintf("生成统计数据失败: %v", err))
		}
	}

	return nil
}

// categorizeProxies 将代理按类别分类
func (cs *ConfigSaver) categorizeProxies() {
	for _, result := range cs.results {
		for i := range cs.categories {
			if cs.categories[i].Filter(result) {
				cs.categories[i].Proxies = append(cs.categories[i].Proxies, result.Proxy)
			}
		}
	}
}

// saveCategory 保存单个类别的代理
func (cs *ConfigSaver) saveCategory(category ProxyCategory) error {
	if len(category.Proxies) == 0 {
		slog.Warn(fmt.Sprintf("yaml节点为空，跳过保存: %s, saveMethod: %s", category.Name, config.GlobalConfig.SaveMethod))
		return nil
	}

	if category.Name == "node.yaml" {
		yamlData, err := yaml.Marshal(map[string]any{
			"proxies": category.Proxies,
		})
		if err != nil {
			return fmt.Errorf("序列化yaml %s 失败: %w", category.Name, err)
		}
		if err := cs.saveMethod(yamlData, category.Name); err != nil {
			return fmt.Errorf("保存 %s 失败: %w", category.Name, err)
		}
		return nil
	}

	return nil
}

// chooseSaveMethod 根据配置选择保存方法
func chooseSaveMethod() func([]byte, string) error {
	switch config.GlobalConfig.SaveMethod {
	case "r2":
		if err := method.ValiR2Config(); err != nil {
			return func(b []byte, s string) error { return fmt.Errorf("R2配置不完整: %v", err) }
		}
		return method.UploadToR2Storage
	case "gist":
		if err := method.ValiGistConfig(); err != nil {
			return func(b []byte, s string) error { return fmt.Errorf("Gist配置不完整: %v", err) }
		}
		return method.UploadToGist
	case "webdav":
		if err := method.ValiWebDAVConfig(); err != nil {
			return func(b []byte, s string) error { return fmt.Errorf("WebDAV配置不完整: %v", err) }
		}
		return method.UploadToWebDAV
	case "local":
		return method.SaveToLocal
	case "s3": // New case for MinIO
		if err := method.ValiS3Config(); err != nil {
			return func(b []byte, s string) error { return fmt.Errorf("S3配置不完整: %v", err) }
		}
		return method.UploadToS3
	default:
		return func(b []byte, s string) error {
			return fmt.Errorf("未知的保存方法或其他方法配置错误: %v", config.GlobalConfig.SaveMethod)
		}
	}
}
