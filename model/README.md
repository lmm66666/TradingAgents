# Model 规范

## 通用规定

1. **所有表必须内嵌 `gorm.Model`**，包含以下默认字段：
   - `ID` — 主键，自增
   - `CreatedAt` — 创建时间
   - `UpdatedAt` — 更新时间
   - `DeletedAt` — 软删除时间（可为 null）

2. **日期字段命名统一使用 `date`**，字符串类型，格式示例：
   - `2022-01-22 11:11:11`

3. **每个模型必须显式实现 `TableName()` 函数**，表名规则：
   - 统一加 `t_` 前缀
   - 使用小写下划线命名（snake_case）
   - 示例：`StockKline` → `t_stock_kline`

4. **一张表一个文件**