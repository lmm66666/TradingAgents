# 项目概述
这是一个使用 agent 分析股票数据辅助投资决策的 Go 项目。

# 项目架构

三层分离：
```
utils/        → 纯计算（MA、MACD、KDJ）
strategy/     → 每个策略 = 1 个条件，返回符合要求的日期+评分
business/     → BacktestService 组合多个策略取交集，回测收益
api/          → HTTP 接口层
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
│   └── stock_kline_weekly.go # 周线数据模型（表 t_stock_kline_weekly）
├── data/                    # 数据访问层（Repository）
│   ├── data.go              # Data 入口，管理数据库连接与各模型 Repo
│   ├── stock_kline_daily.go # 日线数据仓库接口与实现
│   └── stock_kline_weekly.go # 周线数据仓库接口与实现
├── business/                # 业务逻辑层
│   ├── stock_service.go     # StockService 接口与实现（聚合、清洗、调用 broker/repo）
│   ├── stock_service_test.go
│   ├── scheduler.go         # 定时任务调度器（每日扫描并补充缺失数据）
│   ├── scheduler_test.go
│   ├── backtest_service.go  # BacktestService 多策略交集扫描与回测
│   └── backtest_service_test.go
├── api/                     # HTTP API 层（gin）
│   ├── api.md               # API 接口文档（含 curl 示例）
│   ├── README.md            # API 规范（一个接口一个文件）
│   ├── router.go            # gin 路由注册
│   ├── handler.go           # 公共 handler 结构体与响应方法
│   ├── handler_test.go      # 公共 mock 与测试工具
│   ├── save_stock_historical_data.go    # POST /api/stocks/historical
│   ├── save_stock_historical_data_test.go
│   ├── get_stock_analysis_data.go       # GET /api/stocks/analysis
│   ├── get_stock_analysis_data_test.go
│   ├── append_stock_data.go             # POST /api/stocks/append
│   ├── append_stock_data_test.go
│   ├── scan_patterns.go                 # POST /api/patterns/scan
│   └── backtest_patterns.go            # GET /api/patterns/backtest
├── pkg/
│   ├── broker/              # 行情数据提供者
│   │   ├── broker.go        # IBroker 统一接口
│   │   ├── sina.go          # SinaBroker（新浪财经实现）
│   │   └── sina_test.go     # 接口测试
│   ├── utils/               # 技术指标计算工具（纯计算，无业务逻辑）
│   │   ├── round.go         # Round4 四舍五入工具
│   │   ├── ma.go            # SMA / EMA / 成交量均线
│   │   ├── macd.go          # MACD 计算
│   │   ├── macd_test.go
│   │   ├── kdj.go           # KDJ 计算
│   │   ├── kdj_test.go
│   │   ├── volume_ma_test.go
│   │   ├── limiter.go
│   │   └── limiter_test.go
│   └── strategy/            # 策略层（每个策略 = 一个条件）
│       ├── strategy.go              # Strategy 接口、Signal 定义、ResolveStrategy
│       ├── volume_surge.go          # 放量上涨+缩量回调条件
│       ├── volume_surge_test.go
│       ├── kdj_oversold.go          # KDJ 超卖条件（J < 阈值）
│       ├── kdj_oversold_test.go
│       ├── ma60_trend.go            # MA60 向上条件
│       ├── ma60_trend_test.go
│       ├── macd_divergence.go       # MACD 背离条件（骨架）
│       └── README.md                # 策略开发规范
└── shell/                   # 脚本工具
    └── save_historical.sh   # 批量保存历史数据脚本
```

# 开发规范
## 强制要求
- 在读取或分析任何目录下的代码时，**如果该目录下存在 README.md，必须先读取 README.md**，以了解该目录的规范、约束和上下文，避免误读代码
- 代码修改完之后调用 code-simplifier 插件，自己检查一下有没有可以优化的地方
- 需求开发完成后，最后一步必须判断是否需要更新 CLAUDE.md, 对应的 README.md, 和代码中相关的注释，保持文档正确

## 代码风格
- 函数不超过 50 行，文件不超过 800 行
- 遵循 Go 标准编码规范

## Git 提交
- 遵循 Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`
- 必须使用英文，禁止中文
- 冒号后必须加空格
- 句尾不加标点
- 一个 commit 等于一件事，多件事拆分到多个 commit 中
- 简短清晰，不超过 50 字符
