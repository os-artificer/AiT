# ait 实现说明

## 包分层

```text
cmd/ait → internal/ait（包名 ait：Run、Execute、help）
         → internal/ait/router     （顶层 argv：Loop、agent 管理、插件、shell）
         → internal/ait/loop       （本机 Agent Loop）
         → internal/ait/dispatch   （经 tooldispatch 执行插件）
         → internal/ait/shell      （工具箱 REPL：install / update / upgrade）
         → internal/ait/agent      （管理 CLI：connect、model、provider、ollama、config）
         → internal/pkg/aitplugin  （manifest 与 Discover）
         → internal/pkg/aitrelease （GitHub Release 拉取与安装）
         → internal/pkg/agentwire  （Loop 配置的单一事实来源）
         → internal/pkg/tooldispatch （共享的插件执行）
         → internal/pkg/pluginbin  （ExecPath 与 Runner）
```

- `internal/ait` 及其子包不得为了执行而直接导入 `internal/ssh` / `ailog` / `aiimg` / `aihost`（只经 `dispatch` / `tooldispatch` 执行）。
- `aitplugin` / `pluginbin` 不得导入 `internal/ait`。
- `agentcore` 不得导入产品包。

## 进程入口

`Execute` / `agent` CLI 入口先经 `config.StripAIExperimentalFlag` 剥离 `--ai-experimental`（见则进程内开启实验门禁，不写 config），再交给 `router.Route`。

## 顶层路由（`internal/ait/router`）

| argv | 路由 |
|------|------|
| （空） | `loop.Run` — Agent Loop |
| 单独 `version` / `-V` / `--version` | 打印版本行 |
| 单独 `help` / `-h` / `--help` | `WriteRootHelp`（仅单 token；`ait help ssh` 等作 Loop 目标） |
| `agent`（无后续，可带 `--plain`） | `loop.Run` |
| `agent` + 管理动词 | `agent.RunAdmin`（`connect` / `model` / `provider` / `ollama` / `config`） |
| `agent <destination>` | `RunAdmin` → `connect`（`LooksLikeConnectDestination`） |
| `agent` + 其他文本 | `loop.Run`，可带 `--plain`，其余为初始目标 |
| `shell`（无后续参数） | 工具箱 REPL（`shell.RunREPL`） |
| `shell -h` / `shell --help` | `WriteShellHelp` |
| `shell` 带其他参数 | `unknown ait shell command`，提示使用 `ait shell` |
| `ssh`、`log`、`img`、`host`、`keep` 等 | 可解析时 `dispatch.RunPlugin`（遗留别名经 `ResolvePluginName` / `LegacyPluginNames`） |
| 其他首词 | `loop.Run`，拼接 argv 为初始目标（例如 `ait list files`；顶层 `ait install` 亦进 Loop） |

## 关键类型

| 部件 | 作用 |
|------|------|
| `aitplugin.Manifest` | v1 JSON schema（`apiVersion`、`name`、`binary`、仅用于帮助的 `flags`） |
| `aitplugin.Discover` | 用 `sync.Once` 扫描 `InstallDir/plugins/<manifest.name>/`（例如 `plugins/ssh/aitssh`），并用 PATH 补齐内置插件 |
| `tooldispatch.Exec` | 查找插件并经 `pluginbin.ExecPath` 执行（stdin / stdout / stderr） |
| `shell.ExecuteLine` | 执行一行 REPL；吞掉裸的 `*exec.ExitError`（AT14） |
| `aitrelease.Execute` | 从 GitHub Release 做 install / update / upgrade |
| `agentwire.LocalLoopConfig` | TUI / plain、超时、Loop 使用的 SSH CLI 运行器（`aitssh`） |
| `agent.RunAdmin` | connect / model / provider / ollama / config 的 Cobra 树 |

## 配置范围

- **`ait agent config`** — `ai.*` 键（`config.ValidateAI`、`EnsureAIDefaults`）；含 `init` / `view` / `edit` / `validate` / `set` / `get`。
- **`aitssh config`**（插件 `ssh`）— CLI 的 `set` / `get` 只接受 `session.*`，并拒绝 `ai.*`（`config.ValidateSSH`）。`aitssh config init` 在文件尚不存在时仍可能经 `normalizeAI` 写出默认 `ai:` 段；`ai.*` 字段应由 `ait agent config` 维护。

## Release 安装（`aitrelease`）

- 安装根目录：`aitplugin.PrimaryInstallDir()`。
- 资源名：`ait-<tag>-<goos>-<arch>.tar.gz`（Windows 为 `.zip`）。
- 测试可注入 `Options.BaseURL` 或环境变量 `AIT_RELEASE_TEST_BASE_URL`。
- 回环夹具：[`mock/testrelease`](../../mock/testrelease)；长驻守护进程：`go run ./mock/aitrelease-mock`。

## Manifest 钩子

每个已发布工具的 `ExecuteWriters` 在默认路由之前处理 `ait-plugin-manifest`，并经 `aitplugin.WriteManifest` 写出内嵌 JSON。

## 相关文档

- CLI：[`../guide/ait.md`](../guide/ait.md)
- Loop 内核：`internal/pkg/agentcore`
