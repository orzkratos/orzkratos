# orzkratos

**简化您的 Kratos 开发工作流**

两个核心工具来改进 Kratos 开发。

## 英文文档

[ENGLISH README](README.md)

## 🚀 核心功能

1. **简化 Proto 文件添加** - 跳过长路径，直接运行 `orzkratos-add-proto demo`
2. **自动服务同步** - 修改 proto 文件时，服务代码自动更新

## 安装

```bash
go install github.com/orzkratos/orzkratos/cmd/orzkratos-add-proto@latest
go install github.com/orzkratos/orzkratos/cmd/orzkratos-srv-proto@latest
```

## ⚠️ 重要安全说明

**开发者说明：** 我构建这些工具是为了让自己的 Kratos 开发更容易，并决定分享给大家。由于它们会修改代码文件，请谨慎使用！

**首次使用者：** 先创建一个演示 Kratos 项目来练习这些命令，熟悉工作流程后再在实际项目中使用。

**Git 用户：** 运行 `orzkratos-srv-proto` 命令前务必提交代码。此命令会自动修改服务代码，所以运行前一定要提交！！

```bash
# git 项目推荐工作流
git add . && git commit -m "Before orzkratos sync"
orzkratos-srv-proto -auto
git diff  # 检查修改内容
```

## 快速开始

### 1. 添加 Proto 文件（简单方式）

**Kratos 方式：**
```bash
cd your-project-root
kratos proto add api/helloworld/demo.proto
```

**使用 orzkratos（更简单）：**
```bash
cd api/helloworld
orzkratos-add-proto -name demo.proto
```

**更简单：**
```bash
cd api/helloworld
orzkratos-add-proto demo.proto
```

**更简单：**
```bash
cd api/helloworld
orzkratos-add-proto demo    # 自动添加 .proto 扩展名
```

**更简单：**
```bash
cd api/helloworld
orzkratos-add-proto    # 自动创建 helloworld.proto
```

### 2. 自动同步服务与 Proto 变更

修改 proto 文件时，保持服务同步：

**同步特定 proto：**
```bash
cd demo-project
orzkratos-srv-proto -name demo.proto
```

**更简单：**
```bash
cd demo-project
orzkratos-srv-proto demo.proto
```

**更简单：同步所有 proto（带确认）：**
```bash
cd demo-project
orzkratos-srv-proto
```

**更简单：自动确认模式（脚本完美选择）：**
```bash
cd demo-project
orzkratos-srv-proto -auto
```

**执行效果：**
- ✅ 新方法添加到您的服务中
- ✅ 删除的方法变为非导出（无编译错误）
- ✅ 方法顺序匹配 proto 定义
- ✅ 您的现有代码保持不变

## 工具功能

### Proto 添加
- 自动检测项目结构
- 无需记住长路径如 `api/helloworld/demo.proto`
- `cd` 到想要放置 proto 的位置并运行命令
- 与 GoLand 的"在终端中打开"功能配合很好 - 右键点击目标 DIR 并输入命令 `orzkratos-add-proto`

### 服务同步
- 读取 `.proto` 文件以理解服务定义
- 与现有 Go 服务实现比较
- 添加缺失方法的正确签名
- 将删除的方法转换为非导出（防止编译错误）
- 维护您的业务逻辑 - 仅更新方法签名

## 💡 使用说明

**📝 注意：** 这些工具旨在简化 Kratos 开发工作流。请谨慎使用任何修改源代码的工具。

**⚠️ 重要：** 运行同步操作前务必提交/备份您的代码！
