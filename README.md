# Trading

一个基于 Go 的股票数据收集与分析平台，从新浪财经获取 A 股行情与财报数据，提供技术指标计算、策略扫描和 HTTP API 查询能力，辅助投资决策。

## 功能

- **行情数据**：拉取并保存指定股票的日线/周线历史 K 线数据
- **财报数据**：拉取并保存股票季度财报核心指标（利润表、盈利能力、偿债能力、运营效率、现金流等）
- **增量更新**：对比数据库已有数据，只补充缺失的部分，避免重复拉取
- **策略扫描**：基于技术指标（MA、MACD、KDJ、成交量）扫描全市场，发现潜在买点
- **HTTP API**：通过 RESTful 接口查询股价、财报、触发数据同步和策略扫描
- **批量脚本**：通过 Shell 脚本批量保存多只股票的历史数据

## 技术栈

- Go 1.25+
- [Gin](https://gin-gonic.com/) Web 框架
- [GORM](https://gorm.io/) + MySQL
- [goccy/go-yaml](https://github.com/goccy/go-yaml) 配置解析
- 新浪财经 API（行情 + 财报）

## 环境要求

- Go 1.25.7 或更高版本
- MySQL 5.7+ 或 8.0+
- curl（用于批量脚本）

## 快速开始

### 1. 克隆项目

```bash
git clone <repo-url> trading
cd trading
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 创建数据库

```bash
mysql -u root -p -e "CREATE DATABASE trading CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

### 4. 配置数据库连接

复制 `config.yaml` 并根据实际情况修改：

```yaml
Config:
  DB:
    Host: 127.0.0.1
    Port: 3306
    User: root
    Password: your_password
    DBName: trading
```

### 5. 启动服务

```bash
go run .
```

服务启动后会自动执行数据库迁移（`AutoMigrate`），创建所需的数据表：

- `t_stock_kline_daily` — 日线 K 线数据
- `t_stock_kline_weekly` — 周线 K 线数据
- `t_financial_reports` — 季度财报数据

默认监听 `http://localhost:8080`。

## API 概览

所有接口均采用统一 JSON 响应格式：

```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

### 数据写入

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/stocks/historical` | 保存单只股票历史 K 线（5年日线 + 1年周线） |
| POST | `/api/stocks/append` | 触发全量股票数据增量补全（异步） |
| POST | `/api/stocks/financial-report` | 保存单只股票近5年财报（20份季度） |
| POST | `/api/stocks/financial-report/append` | 触发全量财报增量补全（异步） |

### 数据查询

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/stocks/price?code=xxx&cycle=daily&pagesize=20&pagenum=1` | 查询股价 K 线（支持日线/周线分页） |
| GET | `/api/stocks/financial-report?code=xxx&pagesize=20&pagenum=1` | 查询季度财报（支持分页） |
| GET | `/api/stocks/signal?strategy=daily_b1_buy` | 按策略扫描买点 |

详细接口文档见 [api/api.md](api/api.md)。

## 批量保存脚本

`shell/` 目录下提供批量保存脚本，用于一次性拉取多只股票的数据：

```bash
# 批量保存历史 K 线
./shell/save_stock_historical.sh

# 批量保存财报数据
./shell/save_financial_report.sh
```

脚本中的股票代码列表定义在 `shell/code/上海.txt` 中，可按需修改。

## 项目结构

```
trading/
├── main.go                  # 程序入口
├── config.yaml              # 应用配置
├── config/                  # 配置定义
├── model/                   # 数据模型（GORM）
│   ├── stock_kline.go
│   ├── stock_kline_daily.go
│   ├── stock_kline_weekly.go
│   └── financial_report.go
├── data/                    # Repository 数据访问层
│   ├── data.go
│   ├── stock_kline_daily.go
│   ├── stock_kline_weekly.go
│   └── financial_report.go
├── business/                # 业务逻辑层
│   ├── stock_service.go     # 数据拉取与保存
│   ├── analysis_service.go  # 策略扫描与数据查询
│   ├── scheduler.go         # 股票数据定时调度
│   └── financial_scheduler.go # 财报数据定时调度
├── api/                     # HTTP API 层（gin）
│   ├── router.go
│   ├── handler.go
│   ├── api.md               # 接口文档
│   └── *.go / *_test.go     # 各接口 handler
├── pkg/
│   ├── broker/              # 行情数据提供者（新浪财经）
│   ├── indicator/           # 技术指标计算（MA/MACD/KDJ）
│   ├── filter/              # 技术指标过滤器
│   └── strategy/            # 策略层（组合 filter）
└── shell/                   # 批量脚本
    ├── save_stock_historical.sh
    ├── save_financial_report.sh
    └── code/上海.txt
```

## 开发

```bash
# 运行测试
go test ./...

# 构建二进制
go build -o trading .

# 运行
go run .
```
