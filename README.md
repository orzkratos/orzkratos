[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/orzkratos/orzkratos/release.yml?branch=main&label=BUILD)](https://github.com/orzkratos/orzkratos/actions/workflows/release.yml?query=branch%3Amain)
[![GoDoc](https://pkg.go.dev/badge/github.com/orzkratos/orzkratos)](https://pkg.go.dev/github.com/orzkratos/orzkratos)
[![Coverage Status](https://img.shields.io/coveralls/github/orzkratos/orzkratos/main.svg)](https://coveralls.io/github/orzkratos/orzkratos?branch=main)
[![Supported Go Versions](https://img.shields.io/badge/Go-1.25+-lightgrey.svg)](https://go.dev/)
[![GitHub Release](https://img.shields.io/github/release/orzkratos/orzkratos.svg)](https://github.com/orzkratos/orzkratos/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/orzkratos/orzkratos)](https://goreportcard.com/report/github.com/orzkratos/orzkratos)

# orzkratos

**Streamline the Kratos development workflow**

Two apps to boost Kratos development.

---

<!-- TEMPLATE (EN) BEGIN: LANGUAGE NAVIGATION -->
## CHINESE README

[中文说明](README.zh.md)
<!-- TEMPLATE (EN) END: LANGUAGE NAVIGATION -->

## Installation

```bash
go install github.com/orzkratos/orzkratos/cmd/orzkratos-add-proto@latest
go install github.com/orzkratos/orzkratos/cmd/orzkratos-srv-proto@latest
```

## ⚠️ Safe Usage Notes

**Note:** I built these apps to make Kratos development fast and decided to share them. Since these apps can change code files, please use with caution!

**New users:** Create a demo Kratos project to practice these commands and get used to the workflow before using on production projects.

**Git users:** Commit code before running `orzkratos-srv-proto` commands. This command auto modifies service code, so commit before running it!!

```bash
# Recommended workflow for git projects
git add . && git commit -m "Before orzkratos sync"
orzkratos-srv-proto -auto
git diff  # Review what changed
```

---

## App 1: orzkratos-add-proto

**Proto File Addition Made Fast** - Skip the long paths, just run `orzkratos-add-proto demo`

### Usage

**Kratos approach:**

```bash
cd your-project-root
kratos proto add api/helloworld/demo.proto
```

**With orzkratos (much more concise):**

```bash
cd api/helloworld
orzkratos-add-proto -name demo.proto
```

**More concise:**

```bash
cd api/helloworld
orzkratos-add-proto demo.proto
```

**Most concise:**

```bash
cd api/helloworld
orzkratos-add-proto demo    # auto-adds .proto extension
```

**Zero-arg mode:**

```bash
cd api/helloworld
orzkratos-add-proto    # auto creates helloworld.proto
```

### Command Line Options

| Option  | Description            | Example                         |
|---------|------------------------|---------------------------------|
| `-name` | Specify proto filename | `-name demo.proto`              |
| (args)  | Proto filename as arg  | `demo.proto` / `demo`           |
| (none)  | Use current DIR name   | auto creates `helloworld.proto` |

### Main Capabilities

- Auto detects project structure
- No need to memorize long paths like `api/helloworld/demo.proto`
- `cd` to the target location and run the command
- Works with GoLand's "Open in Terminal" feature - right-click the target DIR and input the command `orzkratos-add-proto`

---

## App 2: orzkratos-srv-proto

**Auto Service Sync** - When proto files change, service code auto updates

### Usage

**Sync specific proto:**

```bash
cd demo-project
orzkratos-srv-proto -name demo.proto
```

**More concise:**

```bash
cd demo-project
orzkratos-srv-proto demo.proto
```

**Sync each proto (with confirmation):**

```bash
cd demo-project
orzkratos-srv-proto
```

**Auto-confirm mode (perfect for scripts):**

```bash
cd demo-project
orzkratos-srv-proto -auto
```

**Mask mode (default, flexible naming):**

```bash
cd demo-project
orzkratos-srv-proto -mask
orzkratos-srv-proto -auto -mask
```

**Disable mask mode (strict naming):**

```bash
cd demo-project
orzkratos-srv-proto -mask=false
orzkratos-srv-proto -auto -mask=false
```

### Command Line Options

| Option  | Description               | Example                  |
|---------|---------------------------|--------------------------|
| `-name` | Specify proto filename    | `-name demo.proto`       |
| (args)  | Proto filename as arg     | `demo.proto`             |
| `-auto` | Skip confirmation prompts | `-auto`                  |
| `-mask` | Mask mode (default: true) | `-mask=false` to disable |

### Sync Features

| Feature            | Description                                         |
|--------------------|-----------------------------------------------------|
| **Add Methods**    | New proto methods auto added to service             |
| **Delete Methods** | Removed proto methods become unexported (lowercase) |
| **Sort Methods**   | Method sequence matches proto definition            |
| **Preserve Code**  | Existing business logic stays intact                |

### Mask Mode (`-mask`)

In non-mask mode, it matches service files via filename (e.g., `greeter.proto` → `greeter.go`).

With `-mask` flag (the default), file/struct names are just "masks" - it checks the embedded `Unimplemented*Server` type instead:

```go
type CustomGreetingHandler struct {
    v1.UnimplementedGreeterServer // <- matched by this
    uc *biz.GreeterUsecase
}
```

**Mask Mode Benefits:** File and struct names can be anything - no naming restrictions.

**Example:**

Default mode requires:

- `greeter.proto` → `greeter.go`
- Struct named `GreeterService`

Mask mode allows renaming service file-name and struct-name based on preference:

When the filename `service/greeter.go` doesn't match the preferred aesthetic, and likewise when it cannot express the complete business meaning of this service,
you can rename `service/greeter.go` to `service/custom_greet_service.go`.
When the struct name `GreeterService` doesn't fit the aesthetics, and likewise when it cannot express the complete function scope of the implementation,
you can also rename the struct `GreeterService` to `CustomGreetService`.

Regardless, the mask mode can auto detect the service via the embedded `v1.UnimplementedGreeterServer` type, achieving auto sync of service code alongside proto changes.

**Tip:** Once using `-mask`, stick with it to keep naming stable.

---

## Mechanism

### Proto Addition App

1. Detects the current location in project structure
2. Calculates the path from project root
3. Builds the complete proto path
4. Executes `kratos proto add` with correct arguments

### Service Sync App

1. Reads the `.proto` files to understand service definitions
2. Generates new service code from proto (to staging DIR)
3. Compares with existing Go service implementations
4. Adds missing methods with correct signatures
5. Converts deleted methods to unexported (prevents compile issues)
6. Sorts methods to match proto definition sequence
7. Keeps business logic intact - updates method signatures

---

## 💡 Usage Notes

**📝 Note:** These apps are designed to streamline Kratos development workflows. Use caution with apps that can change source code.

**⚠️ Caution:** Commit/backup code before running sync operations!

---

<!-- TEMPLATE (EN) BEGIN: STANDARD PROJECT FOOTER -->
<!-- VERSION 2025-11-25 03:52:28.131064 +0000 UTC -->

## 📄 License

MIT License - see [LICENSE](LICENSE).

---

## 💬 Contact & Feedback

Contributions are welcome! Report bugs, suggest features, and contribute code:

- 🐛 **Mistake reports?** Open an issue on GitHub with reproduction steps
- 💡 **Fresh ideas?** Create an issue to discuss
- 📖 **Documentation confusing?** Report it so we can improve
- 🚀 **Need new features?** Share the use cases to help us understand requirements
- ⚡ **Performance issue?** Help us optimize through reporting slow operations
- 🔧 **Configuration problem?** Ask questions about complex setups
- 📢 **Follow project progress?** Watch the repo to get new releases and features
- 🌟 **Success stories?** Share how this package improved the workflow
- 💬 **Feedback?** We welcome suggestions and comments

---

## 🔧 Development

New code contributions, follow this process:

1. **Fork**: Fork the repo on GitHub (using the webpage UI).
2. **Clone**: Clone the forked project (`git clone https://github.com/yourname/repo-name.git`).
3. **Navigate**: Navigate to the cloned project (`cd repo-name`)
4. **Branch**: Create a feature branch (`git checkout -b feature/xxx`).
5. **Code**: Implement the changes with comprehensive tests
6. **Testing**: (Golang project) Ensure tests pass (`go test ./...`) and follow Go code style conventions
7. **Documentation**: Update documentation to support client-facing changes
8. **Stage**: Stage changes (`git add .`)
9. **Commit**: Commit changes (`git commit -m "Add feature xxx"`) ensuring backward compatible code
10. **Push**: Push to the branch (`git push origin feature/xxx`).
11. **PR**: Open a merge request on GitHub (on the GitHub webpage) with detailed description.

Please ensure tests pass and include relevant documentation updates.

---

## 🌟 Support

Welcome to contribute to this project via submitting merge requests and reporting issues.

**Project Support:**

- ⭐ **Give GitHub stars** if this project helps you
- 🤝 **Share with teammates** and (golang) programming friends
- 📝 **Write tech blogs** about development tools and workflows - we provide content writing support
- 🌟 **Join the ecosystem** - committed to supporting open source and the (golang) development scene

**Have Fun Coding with this package!** 🎉🎉🎉

<!-- TEMPLATE (EN) END: STANDARD PROJECT FOOTER -->

---

## GitHub Stars

[![Stargazers](https://starchart.cc/orzkratos/orzkratos.svg?variant=adaptive)](https://starchart.cc/orzkratos/orzkratos)
