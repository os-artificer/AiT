# 贡献指南

**语言：** [English](CONTRIBUTING.md) · 简体中文（本文）

感谢参与 **AiT**（AI Terminal）开发。本文档汇总本仓库的**开发流程**与 **Go 编码规范**，供人类贡献者查阅。

- 模块路径：`github.com/os-artificer/AiT`
- 需要 **Go 1.25+**（见 [`go.mod`](go.mod)）
- 提交 PR 前请运行合并门禁 [`.cursor/rules/change-acceptance.mdc`](.cursor/rules/change-acceptance.mdc)：`make fmt`、`make vet`、`make test`、`make build`、`make test-cross`

规范以 [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) 与 [`.cursor/rules/code-reuse.mdc`](.cursor/rules/code-reuse.mdc) 为权威来源；**若更新编码约定，请同步修改上述规则与中英文 CONTRIBUTING 文档**（[`CONTRIBUTING.md`](CONTRIBUTING.md)、[`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md)）。

CLI 命令的 `--help` 须与 `ait login --help` 同版式，正文在 [`internal/ait/helpdoc`](internal/ait/helpdoc)；见 [`.cursor/rules/cli-help.mdc`](.cursor/rules/cli-help.mdc)。

---

## 开发环境

```bash
make build       # 编译到 bin/ait
make test        # go test ./...
make vet         # go vet ./...
make fmt         # go fmt ./...
make test-cross  # Windows / macOS 交叉编译门禁
./bin/ait        # 运行二进制
```

Windows 交叉编译示例：

```bash
GOOS=windows GOARCH=amd64 go build -o bin/ait.exe ./cmd/ait
```

---

## 仓库布局

AiT 是 Go CLI，目录遵循 [golang-standards/project-layout](https://github.com/golang-standards/project-layout)。

| 路径 | 用途 |
|------|------|
| [`cmd/ait`](cmd/ait) | 主程序入口；**仅**保留极简 `main` |
| [`internal/ait`](internal/ait) | 应用入口（`Run`、Cobra CLI、`ait shell` REPL 等） |
| [`internal/pkg/*`](internal/pkg) | 实现库：`store`、`session`、`ssh`、`ai`、`transfer`、`config`、`creds` 等 |
| [`pkg/`](pkg/) | 仅供**外部模块**导入的库（当前为空） |
| [`configs/`](configs/) | 配置模板（如 `ait.example.yaml`） |

约定：

- 优先在 `internal/ait` 与 `internal/pkg/*` 中扩展功能，避免随意新增顶层目录。
- `internal/` 为模块私有，**不得**被本模块以外的代码 import。
- 各目录细节见 [`AGENTS.zh-CN.md`](AGENTS.zh-CN.md)；产品架构见 [`docs/design.zh-CN.md`](docs/design.zh-CN.md)。
- 子目录 `api/`、`deployments/` 等若有 `README.md`，请遵循其说明。

---

## Go 编码规范

生成或重构 Go 代码时**必须**遵守以下约定。

### 总则：可读性优先

全部约定以**可读性为最高优先级**。与其它条款冲突时，默认选择更易读的写法。

仅在确有性能瓶颈、且优化会牺牲可读性时例外；须在改动处用简短注释说明性能理由。禁止为炫技牺牲清晰度。

### 1. import 顺序

三组，组间各空一行：

1. 标准库
2. 本项目包（`github.com/os-artificer/AiT/...`）
3. 第三方包

运行 `make fmt` / `goimports` 后，确认组间空行仍在。

```go
import (
	"context"
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/store"

	"github.com/spf13/cobra"
)
```

### 2. 文件内声明顺序

自上而下：

1. 包注释（如有）
2. 文件级 `const` / `var`（全局常量、变量）
3. 带方法的 `struct` 定义
4. 该 struct 的方法（见第 3 条可见性排序）

### 3. 函数按可见性排序

- **包级函数**：导出（大写开头）在前，未导出在后。
- **方法**：同一 receiver 上，导出方法在前，未导出在后。

### 4. 非 struct 函数独立文件

无 receiver 的包级函数放在独立文件中（如 `*_helpers.go`、`*_parse.go`），仍遵守第 3 条排序。

**例外**：[`cmd/ait/main.go`](cmd/ait/main.go) 仅保留极简 `main`。

包级辅助函数不要与 struct 方法混在同一文件；参考 `internal/pkg/ssh/client` 中的 `*_helpers.go` 布局。

### 5. 禁止硬编码

魔法字符串、数字、重复字面量须提为命名常量：

- **多文件共享**：包内统一文件（如 `constants.go`）或该子包约定的单一常量文件。
- **仅本文件使用**：放在该文件首部 `const` / `var` 块（第 2 条顺序之前或与文件级声明一起）。

### 6. 单 struct 单文件

每个 `.go` 文件最多包含**一个**带方法定义的 struct。

纯数据结构（无方法，或仅作 JSON/SQL/配置字段承载）不受限；可集中在 `types.go` 或与相关逻辑同包分文件。

### 7. 导出函数注释

所有包级**导出**函数、方法须有完整 [godoc](https://go.dev/doc/comment)：

- 功能说明
- 每个参数的含义与限制
- 返回值说明
- 若返回 `error`，注释中列出**可能**的错误条件（类型或场景）

```go
// Login starts an SSH session for dest.
// host is the SSH destination; alias, if non-empty, must be unique in the store.
// Returns store.ErrNotFound if alias is unknown, or a dial/auth error from ssh/client.
func Login(ctx context.Context, host, alias string) (*Session, error) { ... }
```

### 8. 行宽

- 单行（含注释）不超过 **150** 个字符。
- 超长时：拆行、提取变量或辅助函数、缩短命名；避免无意义换行。
- `go fmt` 不检查行宽；生成/重构后需目视或 IDE 标尺确认（见 [`.editorconfig`](.editorconfig) 中 `max_line_length = 150`）。

### 9. 嵌套层级

- 控制流（`if` / `for` / `switch` / `select`）嵌套**最多 2 层**；第 3 层及以上须重构。
- **如非必要**不超 2 层：协议解析、状态机等确实需要更深嵌套时允许，须加一行注释说明为何不能扁平化。
- 优先：提前 `return` / `continue` / `break`、卫语句（guard clause）、提取未导出函数（遵守第 4 条独立文件约定）。

```go
// 不推荐：3 层嵌套
for _, s := range sessions {
	if s.Active {
		if s.Tmux {
			attach(s)
		}
	}
}

// 推荐：卫语句，至多 2 层
for _, s := range sessions {
	if !s.Active || !s.Tmux {
		continue
	}
	attach(s)
}
```

### 10. 可读性优先（细则）

- 与其它条款冲突时，默认选**更易读**的写法（见总则）。
- 性能优化（内联、减少分配、合并循环等）仅在 profiling 或明确热点路径上采用；改动处注释性能理由。
- 禁止：过度一行写法、晦涩缩写、无注释的深嵌套。

### 11. 函数内职责块空行

同一函数内，按**职责阶段**分块，块与块之间保留**恰好一行**空行，便于扫读。

**典型分块**：

- 入参校验 / 默认值
- 资源获取（open、dial、NewSession）
- 核心业务逻辑
- 清理（defer、关闭连接）

**不必**额外空行：

- 紧密相关的 2–3 行（如 `err := f(); if err != nil { return err }`）
- 同一 `switch` / `case` 内连续赋值
- 函数体极短（约 ≤5 行）且仅单一职责

`go fmt` 不会自动插入职责空行；生成/重构后需目视检查。

```go
func Load(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("empty id")
	}

	sess, err := m.store.GetSession(id)
	if err != nil {
		return err
	}

	client, err := m.dial(sess)
	if err != nil {
		return err
	}
	defer client.Close()

	return m.attach(ctx, client, sess)
}
```

---

## 格式化与编辑器

项目使用 [`.editorconfig`](.editorconfig)：

| 范围 | 约定 |
|------|------|
| Go、`Makefile`、`go.mod` | `indent_style = tab`，`max_line_length = 150` |
| YAML、JSON | 2 空格缩进 |
| 通用 | UTF-8、LF、文件末尾换行、去除行尾空格（Markdown 除外） |

Go 修改完成后运行 `make fmt`。`make fmt` **不会**自动保证行宽 ≤150 或函数内职责空行，提交前请人工检查。

---

## 测试

- 测试文件 `*_test.go` 与实现放在同一目录。
- `make test` 必须通过。
- SQLite store 测试优先使用 `:memory:` 数据库。
- 仅在有意义的场景下新增测试；避免只断言显而易见行为的测试。

---

## 依赖与安全

- **依赖**：按需添加，见 [`go.mod`](go.mod)；避免无关的大型依赖。
- **密钥**：API Key、SSH 密码等存放在 OS keyring（[`internal/pkg/creds`](internal/pkg/creds)）。
- **禁止**：将密码写入 SQLite；在日志中输出密钥；提交 `.env` 或凭据文件。
- **开发**：仅在开发环境使用 `AIT_SSH_INSECURE=1` 等不安全选项（见 [`AGENTS.zh-CN.md`](AGENTS.zh-CN.md) 注意事项）。

---

## 提交前检查清单

- [ ] `make fmt`
- [ ] `make vet`
- [ ] `make test`（跨平台变更建议再跑 `make test-cross`）
- [ ] 新/改 Go 代码符合上文 **Go 编码规范**（含行宽、嵌套、godoc、职责空行）
- [ ] 无与 PR 无关的 drive-by 重构
- [ ] 未提交密钥、`.env` 或本地凭据

---

## 延伸阅读

| 文档 | 内容 |
|------|------|
| [`CONTRIBUTING.zh-CN.md`](CONTRIBUTING.zh-CN.md) | 本文：贡献流程与编码规范 |
| [`CHANGELOG.zh-CN.md`](CHANGELOG.zh-CN.md) | 产品与 CLI 更新日志 |
| [`CODE_OF_CONDUCT.zh-CN.md`](CODE_OF_CONDUCT.zh-CN.md) | 行为准则 |
| [`docs/user-manual.zh-CN.md`](docs/user-manual.zh-CN.md) | 用户手册 |
| [`AGENTS.zh-CN.md`](AGENTS.zh-CN.md) | 代码结构与 Agent 导向 |
| [`docs/design.zh-CN.md`](docs/design.zh-CN.md) | 产品架构与行为 |
| [`.cursor/rules/go-structure.mdc`](.cursor/rules/go-structure.mdc) | Go 规范权威来源（Cursor / Agent 同步用） |
| [`.cursor/rules/project.mdc`](.cursor/rules/project.mdc) | 项目范围与构建约定 |
