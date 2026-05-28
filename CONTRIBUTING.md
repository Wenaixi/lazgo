# contributing

感谢你对 lazgo 的兴趣！我们欢迎任何形式的贡献。

## 如何贡献

### 报告问题

请搜索现有 Issue 确认没有重复后，再提交新 Issue。
- Bug 报告：请包含复现步骤和环境信息
- 功能请求：请说明使用场景和预期行为

### Pull Request

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feat/amazing-feature`)
3. 提交更改
4. 推送到 Fork 的仓库
5. 打开 Pull Request

### 开发环境

```bash
git clone https://github.com/lazgo/lazgo.git
cd lazgo/go
go mod download
go test ./...
```

## 代码规范

- 使用 `go fmt` 格式化代码
- 确保测试通过：`go test ./...`
- 添加新功能时请更新对应文档

## 提交信息规范

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `refactor`: 重构
- `test`: 测试
- `chore`: 构建/工具

## 许可证

贡献即表明你的代码将遵循 [MIT License](LICENSE)。