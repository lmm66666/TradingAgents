# 项目概述
这是一个使用 agent 分析股票数据辅助投资决策的 Go 项目。

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
│   └── stock_kline.go       # StockKline CRUD 接口与实现（含 Upsert）
├── business/                # 业务逻辑层
│   ├── stock_service.go     # StockService 接口与实现（聚合、清洗、调用 broker/repo）
│   ├── stock_service_test.go
│   ├── scheduler.go         # 定时任务调度器（每日扫描并补充缺失数据）
│   └── scheduler_test.go
├── api/                     # HTTP API 层（gin）
│   ├── api.md               # API 接口文档（含 curl 示例）
│   ├── router.go            # gin 路由注册
│   ├── stock_handler.go     # HTTP handler（统一 JSON 响应格式）
│   └── stock_handler_test.go
├── pkg/
│   ├── broker/              # 行情数据提供者
│   │   ├── broker.go        # IBroker 统一接口
│   │   ├── sina.go          # SinaBroker（新浪财经实现）
│   │   └── sina_test.go     # 接口测试
│   └── utils/               # 技术指标计算工具
│       ├── indicator.go     # MACD / KDJ 计算
│       └── indicator_test.go
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