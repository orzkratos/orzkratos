[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/orzkratos/orzkratos/release.yml?branch=main&label=BUILD)](https://github.com/orzkratos/orzkratos/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/orzkratos/orzkratos)](https://pkg.go.dev/github.com/orzkratos/orzkratos)
[![Coverage Status](https://img.shields.io/coveralls/github/orzkratos/orzkratos/main.svg)](https://coveralls.io/github/orzkratos/orzkratos?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/orzkratos/orzkratos.svg)](https://github.com/orzkratos/orzkratos/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/orzkratos/orzkratos)](https://goreportcard.com/report/github.com/orzkratos/orzkratos)

# orzkratos

**简化 Kratos 开发工作流**

两个应用来加速 Kratos 开发。

---

<!-- TEMPLATE (ZH) BEGIN: LANGUAGE NAVIGATION -->
## 英文文档

[ENGLISH README](README.md)
<!-- TEMPLATE (ZH) END: LANGUAGE NAVIGATION -->

## 安装

```bash
go install github.com/orzkratos/orzkratos/cmd/orzkratos-add-proto@latest
go install github.com/orzkratos/orzkratos/cmd/orzkratos-srv-proto@latest
```

## ⚠️ 安全使用说明

**说明：** 我构建这些应用是为了让 Kratos 开发更快，并决定分享给大家。由于它们会修改代码文件，请谨慎使用！

**新用户：** 先创建一个演示 Kratos 项目来练习这些命令，熟悉工作流程后再在生产项目中使用。

**Git 用户：** 运行 `orzkratos-srv-proto` 命令前务必提交代码。此命令会自动修改服务代码，所以运行前一定要提交！！

```bash
# git 项目推荐工作流
git add . && git commit -m "Before orzkratos sync"
orzkratos-srv-proto -auto
git diff  # 检查修改内容
```

---

## 应用 1: orzkratos-add-proto

**快速添加 Proto 文件** - 跳过长路径，直接运行 `orzkratos-add-proto demo`

### 使用方式

**Kratos 方式：**

```bash
cd your-project-root
kratos proto add api/helloworld/demo.proto
```

**使用 orzkratos（更简洁）：**

```bash
cd api/helloworld
orzkratos-add-proto -name demo.proto
```

**更简洁：**

```bash
cd api/helloworld
orzkratos-add-proto demo.proto
```

**最简洁：**

```bash
cd api/helloworld
orzkratos-add-proto demo    # 自动添加 .proto 扩展名
```

**零参数模式：**

```bash
cd api/helloworld
orzkratos-add-proto    # 自动创建 helloworld.proto
```

### 命令行选项

| 选项      | 说明            | 示例                      |
|---------|---------------|-------------------------|
| `-name` | 指定 proto 文件名  | `-name demo.proto`      |
| (args)  | proto 文件名作为参数 | `demo.proto` / `demo`   |
| (none)  | 使用当前 DIR 名    | 自动创建 `helloworld.proto` |

### 主要功能

- 自动检测项目结构
- 无需记忆长路径如 `api/helloworld/demo.proto`
- `cd` 到目标位置并运行命令
- 与 GoLand 的"在终端中打开"功能配合使用 - 右键点击目标 DIR 并输入命令 `orzkratos-add-proto`

---

## 应用 2: orzkratos-srv-proto

**自动服务同步** - proto 文件变更时，服务代码自动更新

### 使用方式

**同步特定 proto：**

```bash
cd demo-project
orzkratos-srv-proto -name demo.proto
```

**更简洁：**

```bash
cd demo-project
orzkratos-srv-proto demo.proto
```

**同步全部 proto（带确认）：**

```bash
cd demo-project
orzkratos-srv-proto
```

**自动确认模式（脚本完美选择）：**

```bash
cd demo-project
orzkratos-srv-proto -auto
```

**面具模式（默认，灵活命名）：**

```bash
cd demo-project
orzkratos-srv-proto -mask
orzkratos-srv-proto -auto -mask
```

**禁用面具模式（严格命名）：**

```bash
cd demo-project
orzkratos-srv-proto -mask=false
orzkratos-srv-proto -auto -mask=false
```

### 命令行选项

| 选项      | 说明            | 示例                 |
|---------|---------------|--------------------|
| `-name` | 指定 proto 文件名  | `-name demo.proto` |
| (args)  | proto 文件名作为参数 | `demo.proto`       |
| `-auto` | 跳过确认提示        | `-auto`            |
| `-mask` | 面具模式（默认开启）    | `-mask=false` 禁用   |

### 同步功能

| 功能       | 说明                   |
|----------|----------------------|
| **添加方法** | proto 新增的方法自动添加到服务   |
| **删除方法** | proto 删除的方法变为非导出（小写） |
| **方法排序** | 方法顺序匹配 proto 定义顺序    |
| **保留代码** | 现有的业务逻辑保持不变          |

### 面具模式 (`-mask`)

在非面具模式下，按文件名匹配服务文件（如 `greeter.proto` → `greeter.go`）。

使用 `-mask` 参数时（默认开启），文件名/结构体名只是"面具"，检查的是嵌入的 `Unimplemented*Server` 类型：

```go
type CustomGreetingHandler struct {
    v1.UnimplementedGreeterServer // <- 按这个匹配
    uc *biz.GreeterUsecase
}
```

**面具模式优势：** 文件和结构体可以任意命名 - 无命名限制。

**示例：**

默认模式要求：

- `greeter.proto` → `greeter.go`
- 结构体命名为 `GreeterService`

面具模式允许根据喜好自由重命名服务文件名和结构体名：

比如假如您觉得 `service/greeter.go` 这个文件名不够符合您的审美，或者不能完整表达其业务含义时，
您可以将 `service/greeter.go` 重命名为 `service/custom_greet_service.go`。
再比如假如您觉得 `GreeterService` 这个类型名不够美观，或者不能够完整涵盖其功能含义时，
您也可以将结构体 `GreeterService` 重命名为 `CustomGreetService`。

我们的面具模式依然能够通过嵌入的 `v1.UnimplementedGreeterServer` 类型自动检测服务，实现伴随 proto 修改自动同步 service 层代码的功能。

**建议：** 一旦使用 `-mask`，建议一直使用以保持命名稳定。

---

## 运行机制

### Proto 添加应用

1. 检测当前位置在项目结构中的位置
2. 计算从项目根目录的路径
3. 构建完整的 proto 路径
4. 使用正确的参数执行 `kratos proto add`

### 服务同步应用

1. 读取 `.proto` 文件以理解服务定义
2. 从 proto 生成新的服务代码（到暂存 DIR）
3. 与现有 Go 服务实现比较
4. 添加缺失方法的正确签名
5. 将删除的方法转换为非导出（防止编译问题）
6. 排列方法以匹配 proto 定义顺序
7. 保持业务逻辑不变 - 更新方法签名

---

## 💡 使用说明

**📝 注意：** 这些应用旨在简化 Kratos 开发工作流。请谨慎使用会修改源代码的应用。

**⚠️ 注意：** 运行同步操作前务必提交/备份代码！

---

<!-- TEMPLATE (ZH) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 许可证类型

MIT 许可证 - 详见 [LICENSE](LICENSE)。

---

## 💬 联系与反馈

非常欢迎贡献代码！报告 BUG、建议功能、贡献代码：

- 🐛 **问题报告？** 在 GitHub 上提交问题并附上重现步骤
- 💡 **新颖思路？** 创建 issue 讨论
- 📖 **文档疑惑？** 报告问题，帮助我们完善文档
- 🚀 **需要功能？** 分享使用场景，帮助理解需求
- ⚡ **性能瓶颈？** 报告慢操作，协助解决性能问题
- 🔧 **配置困扰？** 询问复杂设置的相关问题
- 📢 **关注进展？** 关注仓库以获取新版本和功能
- 🌟 **成功案例？** 分享这个包如何改善工作流程
- 💬 **反馈意见？** 欢迎提出建议和意见

---

## 🔧 代码贡献

新代码贡献，请遵循此流程：

1. **Fork**：在 GitHub 上 Fork 仓库（使用网页界面）
2. **克隆**：克隆 Fork 的项目（`git clone https://github.com/yourname/repo-name.git`）
3. **导航**：进入克隆的项目（`cd repo-name`）
4. **分支**：创建功能分支（`git checkout -b feature/xxx`）
5. **编码**：实现您的更改并编写全面的测试
6. **测试**：（Golang 项目）确保测试通过（`go test ./...`）并遵循 Go 代码风格约定
7. **文档**：面向用户的更改需要更新文档
8. **暂存**：暂存更改（`git add .`）
9. **提交**：提交更改（`git commit -m "Add feature xxx"`）确保向后兼容的代码
10. **推送**：推送到分支（`git push origin feature/xxx`）
11. **PR**：在 GitHub 上打开 Merge Request（在 GitHub 网页上）并提供详细描述

请确保测试通过并包含相关的文档更新。

---

## 🌟 项目支持

非常欢迎通过提交 Merge Request 和报告问题来贡献此项目。

**项目支持：**

- ⭐ **给予星标**如果项目对您有帮助
- 🤝 **分享项目**给团队成员和（golang）编程朋友
- 📝 **撰写博客**关于开发工具和工作流程 - 我们提供写作支持
- 🌟 **加入生态** - 致力于支持开源和（golang）开发场景

**祝你用这个包编程愉快！** 🎉🎉🎉

<!-- TEMPLATE (ZH) END: STANDARD PROJECT FOOTER -->

---

## GitHub 标星点赞

[![Stargazers](https://starchart.cc/orzkratos/orzkratos.svg?variant=adaptive)](https://starchart.cc/orzkratos/orzkratos)
