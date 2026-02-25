# Short URL API
![Gemini_Generated_Image_bc7t7kbc7t7kbc7t.png](https://youke.xn--y7xa690gmna.cn/s1/2026/02/21/69988cb029bf6.webp)

基于 Go (Gin + GORM) 和 Redis 实现的高性能短链接生成与跳转服务。采用前后端分离的 RESTful API 设计。

## 核心特性

- **RESTful API**：纯 JSON 接口，完美支持前后端分离架构。
- **工业级发号器**：基于 Redis `INCR` 指令与 Base62 编码算法，保证短码极短且全局唯一，彻底告别哈希冲突。
- **持久化存储**：使用 SQLite (配合 GORM) 进行轻量级的数据落盘。
- **中间件扩展**：内置耗时统计等自定义 Gin 中间件。

## 技术栈

- **Web 框架**: [Gin](https://gin-gonic.com/)
- **ORM 框架**: [GORM](https://gorm.io/) + SQLite
- **缓存与发号器**: [Redis](https://redis.io/) (go-redis/v9)
- **环境要求**: Go 1.20+ , Docker (用于运行 Redis)

## 快速开始

### 1. 环境准备 (启动 Redis)
本项目依赖 Redis 运行发号器。请确保你的机器已安装 Docker，并执行以下命令在后台启动一个 Redis 容器：
```bash
docker run -d --name short-url-redis -p 6379:6379 redis

```

### 2. 克隆与运行

```bash
# 克隆项目
git clone [https://github.com/senorl/short-url.git](https://github.com/senorl/short-url.git)
cd short-url

# 下载依赖
go mod tidy

# 运行服务 (默认监听 8080 端口)
go run cmd/server/main.go

```

## API 接口说明 (使用 Postman 或 Curl 测试)

### 1. 生成短链接

* **路径**: `POST /shorten`
* **请求头**: `Content-Type: application/json`
* **请求体 (Body - raw - JSON)**:
```json
{
    "url": "[https://www.bilibili.com](https://www.bilibili.com)"
}

```


* **成功响应**:
```json
{
    "code": 200,
    "msg": "success",
    "shortcode": "xxx"
}

```



### 2. 访问/跳转短链接

* **路径**: `GET /:shortcode` (例如: `http://localhost:8080/a`)
* **说明**: 在浏览器直接输入该地址，或在 Postman 中访问。系统会自动执行 302 重定向到原始的长链接。

### 3. 获取短链接列表

* **路径**: `GET /api/links`
* **成功响应**: 返回数据库中所有生成的短链接记录数组。

