# Short URL 

本项目是一个基于 Go 语言实现的高性能分布式短链接系统。通过三级缓存架构与异步日志解耦，系统在单机环境下可支撑 20,000+ QPS 的并发响应。
![Gemini_Generated_Image_bc7t7kbc7t7kbc7t.png](https://youke.xn--y7xa690gmna.cn/s1/2026/02/21/69988cb029bf6.webp)

## 核心特性

* **高性能架构**：采用 Local Cache (内存) + Redis + MySQL 三级缓存防御体系，最大限度降低数据库负载。
* **异步日志解耦**：引入 Kafka 处理访问日志，实现业务逻辑与持久化存储的彻底解耦。
* **批量落盘策略**：消费者端采用缓冲区机制，凑够批量数据后单次写入 MySQL，优化 I/O 性能。
* **工业级发号器**：基于 Redis INCR 与 Base62 编码，确保短码全局唯一且生成的复杂度为 O(1)。
* **全方位监控**：集成 pprof 性能分析探针，支持通过火焰图实时诊断系统瓶颈。

## 技术栈

* **后端框架**: Gin
* **数据库**: MySQL + GORM
* **中间件**: Redis (缓存/发号器) + Kafka (消息队列)
* **性能工具**: wrk (压测) + pprof (诊断)

---

## 环境准备与 Docker 部署

项目依赖多个中间件，建议通过以下官方镜像快速启动环境。

### 1. 启动 Redis (发号器与缓存)

```bash
docker run -d --name short-url-redis -p 6379:6379 redis

```

### 2. 启动 MySQL

```bash
docker run -d --name short-url-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=your_password mysql:8.0

```

### 3. 启动 Kafka (异步日志)


```bash
docker run -d --name kafka-server -p 9092:9092 apache/kafka:3.7.0

```

---

## 快速开始

### 1. 克隆与配置

```bash
git clone https://github.com/senorl/short-url.git
cd short-url
go mod tidy

```

### 2. 数据库初始化

在 `cmd/server/main.go` 中配置好数据库连接串后运行，程序将自动通过 GORM 进行 AutoMigrate 建表。

### 3. 运行服务

```bash
go run cmd/server/main.go

```

---

## 接口测试说明

### 1. 生成短链接 (POST)

* **路径**: `/shorten`
* **示例请求**:

```bash
curl -X POST http://localhost:8080/shorten \
     -H "Content-Type: application/json" \
     -d '{"url": "https://github.com"}'

```

### 2. 重定向跳转 (GET)

* **路径**: `/:shortcode`
* **示例**: `http://localhost:8080/VLBH4Z`
---

## 性能压测与诊断

### 1. 极限压测

使用 `wrk` 工具模拟 1000 个并发连接对重定向接口进行 30 秒测试：

```bash
wrk -t12 -c1000 -d30s http://localhost:8080/你的短码

```

### 2. 性能诊断 (pprof)

在压测进行期间，执行以下命令抓取 CPU 性能数据并生成可视化火焰图：

```bash
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile?seconds=30

```

### 3. 表现

在 MacBook M1 Pro 环境下，系统表现如下：

* **GET 接口**: 20,228 QPS (Avg Latency: 61ms)
* **POST 接口**: 17,210 QPS
* **错误率**: < 0.01% (1000 并发下)

---

## 目录结构

* `cmd/server/`: 程序入口
* `internal/api/`: Gin 路由与 Handler 逻辑
* `internal/model/`: 数据库模型定义
* `internal/worker/`: Kafka 消费者批量落盘逻辑
* `benchmark/`: 自动化压测脚本与 Lua 配置

