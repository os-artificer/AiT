# ait — Agent 交互入口

无参 `ait` 或 `ait agent` 进入 Agent Loop。未配置 AI 时退出，stderr 打印 `AI not configured` 与 `Try:`（DeepSeek、Minimax、Ollama、`ait agent --help`）。

顶层可分发插件：`ait ssh`、`ait log`、`ait img`、`ait host`、`ait keep`。`ait shell` 是工具箱 REPL（提示符 `ait>`）；`install` / `update` / `upgrade` 仅在该 REPL 内置，不在顶层 argv。

## 用法

```bash
ait                              # Agent Loop（默认）
ait agent [goal...]              # 同一 Loop；可带初始目标
ait agent --plain [goal...]      # 纯文本界面（仅 ait agent）
ait agent connect user@host      # 远程 Agent Loop
ait agent user@host              # 同上（裸 destination）
ait agent model list             # 列出已配置模型
ait agent --model minimax …      # 本会话覆盖默认提供商
ait agent provider enable deepseek
ait agent ollama status|setup|start|stop|disable
ait agent config init|view|edit|validate|set|get
ait --ai-experimental …          # 进程内开启实验门禁（不写 config）
ait ssh user@host.example.com    # 插件分发（不进 REPL）
ait log search error
ait shell                        # 工具箱 REPL
ait version / ait -V             # 版本
ait -h                           # 根帮助（仅单独 help/-h/--help）
ait shell -h                     # 工具箱 REPL 帮助
```

首词不是保留路由（`agent` / `shell` / `version`）、已发现插件名或单独的 help/version 时，整段 argv 当作 Loop 初始目标（如 `ait list files`）。顶层 `ait install …` 等也会进 Loop，**不会**执行 shell 安装。

## Agent Loop

- TTY 默认 TUI；`ait agent --plain` 为纯文本。
- 只读命令自动执行；写操作与高危命令需确认（`y` / `n` 等）。
- 无 one-shot `ask` / `agent` 子命令；多轮任务用 Loop。
- 远程：`ait agent connect <destination>` 或 `ait agent <destination>`。

关闭实验能力：`ait agent config set ai.experimental.enabled false` 或 `AIT_AI_EXPERIMENTAL=0`。云提供商需 API key；`ai.tiered.enabled` 时短提示可走本机 Ollama（仍需 Ollama 可达）。

## 工具箱 shell（`ait shell`）

```text
ait> help
ait> install
ait> update
ait> upgrade
ait> ssh user@host.example.com
ait> log search error
ait> exit
```

内置：`help`、`exit` / `quit`、`version`、`history`、`clear`、`install`、`update`、`upgrade`。Agent Loop 不从 shell REPL 启动。

## 插件命令与二进制

| 命令 | 二进制 |
|------|--------|
| `ssh` | `aitssh` |
| `log` | `aitlog` |
| `img` | `aitimg` |
| `host` | `aithost` |
| `keep` | `aitkeep` |

遗留名仍映射：`aissh`→`ssh`，`ailog`→`log`，`aiimg`→`img`，`aihost`→`host`；以及二进制名 `aitssh` / `aitlog` / `aitimg` / `aithost` / `aitkeep`。

## install / update / upgrade

仅在 **`ait shell`** 内从 [GitHub Release](https://github.com/os-artificer/AiT/releases) 拉取官方二进制。官方插件：`ssh`、`log`、`img`、`host`、`keep`。

`upgrade` 不接受插件名。无参时更新 `ait` 与已装官方插件。`update` 可带插件名，只更新已装官方插件。

路由与包布局见 [../architecture/ait.md](../architecture/ait.md)。
