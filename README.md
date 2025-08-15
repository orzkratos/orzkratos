# orzkratos

**Simplify your Kratos development workflow**

Two core tools to improve Kratos development.

## CHINESE README

[中文说明](README.zh.md)

## 🚀 Core Features

1. **Simplified Proto Addition** - Skip the long paths, just run `orzkratos-add-proto demo` 
2. **Auto Service Sync** - When you change proto files, service code auto updates

## Install

```bash
go install github.com/orzkratos/orzkratos/cmd/orzkratos-add-proto@latest
go install github.com/orzkratos/orzkratos/cmd/orzkratos-srv-proto@latest
```

## ⚠️ Important Safety Notes

**Developer's Note:** I built these utils to make my own Kratos development easier and decided to share. Since they modify code files, please use with care!

**First time users:** Create a demo Kratos project first to practice these commands and get familiar with the workflow before using on your actual projects.

**Git users:** Always commit your code before running `orzkratos-srv-proto` commands. This command auto modifies service code, so commit before you run it!!

```bash
# Recommended workflow for git projects
git add . && git commit -m "Before orzkratos sync"
orzkratos-srv-proto -auto
git diff  # Review what changed
```

## Quick Start

### 1. Add Proto Files (the easy way)

**Kratos way:**
```bash
cd your-project-root
kratos proto add api/helloworld/demo.proto
```

**With orzkratos (much simpler):**
```bash
cd api/helloworld
orzkratos-add-proto -name demo.proto
```

**Even simpler:**
```bash
cd api/helloworld
orzkratos-add-proto demo.proto
```

**Even simpler:**
```bash
cd api/helloworld
orzkratos-add-proto demo    # auto-adds .proto extension
```

**Even simpler:**
```bash
cd api/helloworld
orzkratos-add-proto    # auto creates helloworld.proto
```

### 2. Auto-Sync Services with Proto Changes

When you modify your proto file, keep services in sync:

**Sync specific proto:**
```bash
cd demo-project
orzkratos-srv-proto -name demo.proto
```

**Even simpler:**
```bash
cd demo-project
orzkratos-srv-proto demo.proto
```

**Even simpler: Sync all protos (with confirmation):**
```bash
cd demo-project
orzkratos-srv-proto
```

**Even simpler: Auto-confirm mode (perfect for scripts):**
```bash
cd demo-project
orzkratos-srv-proto -auto
```

**What happens:**
- ✅ New methods added to your service
- ✅ Deleted methods become unexported (no compile errors)
- ✅ Method order matches proto definition
- ✅ Your existing code stays untouched

## What This Tool Does

### Proto Addition
- Detects your project structure auto
- No need to remember long paths like `api/helloworld/demo.proto`
- `cd` to where you want the proto and run the command
- Works great with GoLand's "Open in Terminal" feature - right-click your target DIR and input the command `orzkratos-add-proto`

### Service Synchronization  
- Reads your `.proto` files to understand service definitions
- Compares with existing Go service implementations
- Adds missing methods with proper signatures
- Converts removed methods to unexported (prevents compile errors)
- Maintains your business logic - only updates method signatures

## 💡 Usage Notes

**📝 Please note:** These tools are designed to simplify Kratos development workflows. Please use caution with any tool that modifies source code.

**⚠️ Important:** Always commit/backup your code before running sync operations!
