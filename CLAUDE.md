# 项目概述
这是一个使用 agent 分析股票数据辅助投资决策的 Go 项目。

# 项目架构

四层分离：
```
pkg/indicator/  → 纯计算（MA、MACD、KDJ）
pkg/filter/     → 技术指标过滤器，输入 K 线返回每天的日期+布尔结果
pkg/strategy/   → 策略层，组合多个 filter 取交集，支持 Scan / ScanAll
business/       → StockService 数据服务、Scheduler 定时任务
api/            → HTTP 接口层
```

# 项目结构

```
trading/
├── go.mod / go.sum          # Go 模块（go 1.25.7，依赖 gin、gorm、x/text）
├── main.go                  # 程序入口（初始化各层并启动 gin server）
├── config.yaml              # 应用配置文件（DB 连接等）
├── .gitignore               # Git 忽略规则
├── config/
│   └── config.go            # 配置结构体定义与加载
├── model/                   # 数据模型层
│   ├── README.md            # Model 规范（进入该目录时必须优先读取）
│   ├── stock_kline.go       # 通用 K 线数据模型（GORM）
│   ├── stock_kline_daily.go # 日线数据模型（表 t_stock_kline_daily）
│   ├── stock_kline_weekly.go # 周线数据模型（表 t_stock_kline_weekly）
│   └── financial_report.go  # 财报数据模型（表 t_financial_reports）
├── data/                    # 数据访问层（Repository）
│   ├── data.go              # Data 入口，管理数据库连接与各模型 Repo
│   ├── stock_kline_daily.go # 日线数据仓库接口与实现
│   ├── stock_kline_weekly.go # 周线数据仓库接口与实现
│   └── financial_report.go  # 财报数据仓库接口与实现
├── business/                # 业务逻辑层
│   ├── stock_service.go     # StockService 接口与实现（数据拉取、清洗、保存）
│   ├── stock_service_test.go
│   ├── analysis_service.go  # AnalysisService 接口与实现（买点扫描、数据查询）
│   ├── scheduler.go         # 股票数据定时任务调度器
│   ├── scheduler_test.go
│   └── financial_scheduler.go # 财报数据定时任务调度器
├── api/                     # HTTP API 层（gin，一个接口一个文件）
│   ├── api.md               # API 接口文档（含 curl 示例）
│   ├── README.md            # API 规范
│   ├── router.go            # gin 路由注册
│   ├── handler.go           # 公共 handler 结构体与响应方法
│   ├── handler_test.go      # 公共 mock 与测试工具
│   ├── save_stock_historical_data.go    # POST /api/stocks/historical
│   ├── save_stock_historical_data_test.go
│   ├── append_stock_data.go             # POST /api/stocks/append
│   ├── append_stock_data_test.go
│   ├── save_financial_report_data.go    # POST /api/stocks/financial-report
│   ├── save_financial_report_data_test.go
│   ├── append_financial_report_data.go  # POST /api/stocks/financial-report/append
│   ├── append_financial_report_data_test.go
│   ├── get_stock_buy_signals.go         # GET /api/stocks/signal
│   ├── get_stock_price.go               # GET /api/stocks/price
│   ├── get_stock_price_test.go
│   ├── get_financial_report.go          # GET /api/stocks/financial-report
│   └── get_financial_report_test.go
├── pkg/
│   ├── broker/              # 行情数据提供者
│   │   ├── broker.go        # IBroker 统一接口
│   │   ├── sina.go          # SinaBroker（新浪财经实现）
│   │   └── sina_test.go     # 接口测试
│   ├── indicator/           # 技术指标计算工具（纯计算，无业务逻辑）
│   │   ├── round.go         # Round4 四舍五入工具
│   │   ├── ma.go            # SMA / EMA / 成交量均线
│   │   ├── macd.go          # MACD 计算
│   │   ├── macd_test.go
│   │   ├── kdj.go           # KDJ 计算
│   │   ├── kdj_test.go
│   │   ├── volume_ma_test.go
│   │   ├── limiter.go
│   │   └── limiter_test.go
│   ├── filter/              # 过滤器层（每个 filter = 一个条件，返回每天 bool）
│   │   ├── filter.go        # IFilter 接口、Result 定义
│   │   ├── filter_test.go
│   │   ├── kdj.go           # KDJ 超买/超卖过滤器
│   │   ├── kdj_test.go
│   │   ├── ma.go            # MA 趋势过滤器
│   │   ├── ma_test.go
│   │   ├── volume_surge.go  # 放量上涨后回调过滤器
│   │   ├── volume_surge_test.go
│   │   ├── date.go          # 持有天数过滤器
│   │   └── date_test.go
│   └── strategy/            # 策略层（组合多个 filter）
│       ├── strategy.go      # Strategy 结构体、Signal、Scan / ScanAll
│       ├── strategy_test.go
│       ├── buy.go           # 预定义买入策略（如 B1）
│       └── sell.go          # 预定义卖出策略
└── shell/                   # 脚本工具
    ├── save_stock_historical.sh   # 批量保存股票历史 K 线数据
    ├── save_financial_report.sh   # 批量保存股票财报数据
    └── code/                      # 股票代码列表
        └── 上海.txt
```

# 开发规范
## 强制要求
- 在读取或分析任何目录下的代码时，**如果该目录下存在 README.md，必须先读取 README.md**，以了解该目录的规范、约束和上下文，避免误读代码
- 代码修改完之后调用 code-simplifier 插件，自己检查一下有没有可以优化的地方
- 需求开发完成后，最后一步必须判断是否需要更新 CLAUDE.md, 对应的 README.md, 和代码中相关的注释，保持文档正确

## 代码风格
- 函数不超过 50 行，文件不超过 800 行
- 遵循 Go 标准编码规范
- 接口名称以 I 打头

## Git 提交
- 遵循 Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`
- 必须使用英文，禁止中文
- 冒号后必须加空格
- 句尾不加标点
- 一个 commit 等于一件事，多件事拆分到多个 commit 中
- 简短清晰，不超过 50 字符
