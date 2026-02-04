# 🚀 订阅检测转换工具

> **✨ 修复逻辑、简化操作、增加功能、节省内存、一键启动无需配置**

> **⚠️ 注意：** 请查看[配置文件](https://github.com/lingyuanzhicheng/subs-check-lite/blob/master/config/config.example.yaml)以了解详细功能配置。

> **❤️ 求饶：** 本仓库的代码修改是随手在Github Dev 上进行修改的，所以有大量垃圾Commits，无需在意。

## 📸 预览

![index](./doc/images/index.png)
| | |
|---|---|
| ![sparkle](./doc/images/sparkle.png)  | ![v2rayn](./doc/images/v2rayn.png)  |
| ![login](./doc/images/login.png)  | ![admin](./doc/images/admin.png)  |
| ![preview](./doc/images/preview.png) | ![dingtalk](./doc/images/dingtalk.png)  |

## ✨ 功能特性

### 更改点

- **移除内置的 Sub-Store**
- **加入 Clash 转 V2ray 订阅**
- **加入用于基本数据展示的主页**
- **调整了管理页面的交互逻辑**
- **默认规则改为个性化规则**
- **调整路由让获取更方便**

### 原项目

- **🔗 订阅合并**
- **🔍 节点可用性检测**
- **🗑️ 节点去重**
- **⏱️ 节点测速**
- **🎬 流媒体平台解锁检测**
- **✏️ 节点重命名**
- **🔄 任意格式订阅转换**
- **🔔 支持100+通知渠道**
- ~~**🌐 内置 Sub-Store**~~
- **🖥️ WEB 控制面板**
- **⏰ 支持 Crontab 表达式**
- **🖥️ 多平台支持**

### 本分支

- **🗑️ 移除了内置的 Sub-Store**
- **↪️ 内置 Clash 转 V2ray 格式**
- **🖥️ 主页数据展示**

## 🛠️ 部署与使用 
> 首次运行会在当前目录生成默认配置文件。

### 🪜 代理设置（可选）
<details>
  <summary>展开查看</summary>

如果拉取非Github订阅速度慢，可使用通用的 HTTP_PROXY HTTPS_PROXY 环境变量加快速度；此变量不会影响节点测试速度
```bash
# HTTP 代理示例
export HTTP_PROXY=http://username:password@192.168.1.1:7890
export HTTPS_PROXY=http://username:password@192.168.1.1:7890

# SOCKS5 代理示例
export HTTP_PROXY=socks5://username:password@192.168.1.1:7890
export HTTPS_PROXY=socks5://username:password@192.168.1.1:7890

# SOCKS5H 代理示例
export HTTP_PROXY=socks5h://username:password@192.168.1.1:7890
export HTTPS_PROXY=socks5h://username:password@192.168.1.1:7890
```
如果想加速github的链接，可使用网上公开的github proxy，或者使用下方自建测速地址处的worker.js自建加速
```
# Github Proxy，获取订阅使用，结尾要带的 /
# github-proxy: "https://ghfast.top/"
github-proxy: "https://custom-domain/raw/"
```

</details>

### 🌐 自建测速地址（可选）
<details>
  <summary>展开查看</summary>

> **⚠️ 注意：** 避免使用 Speedtest 或 Cloudflare 下载链接，因为部分节点会屏蔽测速网站。

1. 将 [worker.js](./doc/cloudflare/worker.js) 部署到 Cloudflare Workers。
2. 绑定自定义域名（避免被节点屏蔽）。
3. 在配置文件中设置 `speed-test-url` 为你的 Workers 地址：

```yaml
# 100MB
speed-test-url: https://custom-domain/speedtest?bytes=104857600
# 1GB
speed-test-url: https://custom-domain/speedtest?bytes=1073741824
```
</details>

### 🐳 Docker 运行

> **⚠️ 注意：**  
> - 限制内存请使用 `--memory="500m"`。  
> - 可通过环境变量 `API_KEY` 设置 Web 控制面板的 API Key。
> - 镜像可用 `moeceo/subs-check:lite` 推送可能会慢一步。

#### 📜 Docker-Compose

```yaml
services:
  subs-check-lite:
    build: .
    image: subs-check:lite
    container_name: subs-check-lite
    network_mode: bridge
    restart: always
    volumes:
      - ./config:/app/config
      - ./output:/app/output
    ports:
      - "8199:8199"
    environment:
      - TZ=Asia/Shanghai
      # - HTTP_PROXY=http://192.168.1.1:7890
      # - HTTPS_PROXY=http://192.168.1.1:7890
      - API_KEY=subs-check-lite
```

### 🖥️ 源码运行

```bash
go run . -f ./config/config.yaml
```

## 🔔 通知渠道配置（可选）
<details>
  <summary>展开查看</summary>

> **📦 支持 100+ 通知渠道**，通过 [Apprise](https://github.com/caronc/apprise) 发送通知。

### 🌐 Vercel 部署

1. 点击[**此处**](https://vercel.com/new/clone?repository-url=https://github.com/beck-8/apprise_vercel)部署 Apprise。
2. 部署后获取 API 链接，如 `https://testapprise-beck8s-projects.vercel.app/notify`。
3. 建议为 Vercel 项目设置自定义域名（国内访问 Vercel 可能受限）。

### 🐳 Docker 部署

> **⚠️ 注意：** 不支持 arm/v7。

```bash
# 基础运行
docker run --name apprise -p 8000:8000 --restart always -d caronc/apprise:latest

# 使用代理运行
docker run --name apprise \
  -p 8000:8000 \
  -e HTTP_PROXY=http://192.168.1.1:7890 \
  -e HTTPS_PROXY=http://192.168.1.1:7890 \
  --restart always \
  -d caronc/apprise:latest
```

### 📝 配置文件中配置通知

```yaml
# 填写搭建的apprise API server 地址
# https://notify.xxxx.us.kg/notify
apprise-api-server: ""
# 填写通知目标
# 支持100+ 个通知渠道，详细格式请参照 https://github.com/caronc/apprise
recipient-url: 
  # telegram格式：tgram://{bot_token}/{chat_id}
  # - tgram://xxxxxx/-1002149239223
  # 钉钉格式：dingtalk://{Secret}@{ApiKey}
  # - dingtalk://xxxxxx@xxxxxxx
# 自定义通知标题
notify-title: "🔔 节点状态更新"
```
</details>

## 💾 保存方法配置

> **⚠️ 注意：** 选择保存方法时，请更改 `save-method` 配置。

- **本地保存**：保存到 `./output` 文件夹。
- **R2**：保存到 Cloudflare R2 [配置方法](./doc/r2.md)。
- **Gist**：保存到 GitHub Gist [配置方法](./doc/gist.md)。
- **WebDAV**：保存到 WebDAV 服务器 [配置方法](./doc/webdav.md)。
- **S3**：保存到 S3 对象存储。

## 📲 订阅使用方法

> **💡 提示：** 项目不内置 Sub-Store 或 Subconverter ，仅提供 Clash 与 V2ray 系订阅

**🚀 通用订阅**
```bash

# Clash 节点订阅
http://ip:port/node

# V2ray 节点订阅
http://ip:port/v2ray

# Clash 规则订阅
http://ip:port/rule

# Clash 节点订阅
http://ip:port/sub/node.yaml

# V2ray 节点订阅
http://ip:port/sub/v2ray.txt

# Clash 规则订阅
http://ip:port/sub/rule.yaml
```

</details>

## 🙏 鸣谢
[cmliu](https://github.com/cmliu)、[beck-8](https://github.com/beck-8/subs-check)、[bestruirui](https://github.com/bestruirui/BestSub)、[iplark](https://iplark.com/)

## ⚖️ 免责声明

本工具仅供学习和研究使用，使用者应自行承担风险并遵守相关法律法规。
