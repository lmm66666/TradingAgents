# 项目概述
这是一个使用 agent 分析股票数据辅助投资决策的 Go 项目。

# 项目结构

```
trading/
├── go.mod / go.sum          # Go 模块（go 1.25.7，依赖 gorm、x/text）
├── main.go                  # 程序入口（当前为占位）
├── model/                   # 数据模型层
│   ├── README.md            # Model 规范（进入该目录时必须优先读取）
│   └── stock_kline.go       # StockKline K线数据模型（GORM）
└── pkg/broker/              # 行情数据提供者
    ├── broker.go            # MarketDataProvider 统一接口
    ├── sina.go              # SinaProvider（新浪财经实现）
    └── sina_test.go         # 接口测试
```

# 开发规范
## 强制要求
- 在读取或分析任何目录下的代码时，**如果该目录下存在 README.md，必须先读取 README.md**，以了解该目录的规范、约束和上下文，避免误读代码。
- 执行代码变更时，最后一步必须判断是否需要更新 CLAUDE.md, 对应的 README.md, 和代码中相关的注释，保持文档正确。

## 代码风格
- 无 emojis 代码/注释/文档
- 优先使用不可变数据
- 函数不超过 50 行，文件不超过 800 行
- 遵循 Go 标准编码规范

## Git 提交
- 遵循 Conventional Commits: `feat:`, `fix:`, `refactor:`, `docs:`, `test:`