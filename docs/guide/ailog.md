# ailog

> **名称：** 二进制 `aitlog`，插件命令 `log`（`ait log …`），Go 包名 `ailog`。

在 Linux、macOS 与 Windows 上从命令行查看与搜索系统日志。

## 构建

```bash
make build   # → bin/aitlog
./bin/aitlog --help
```

## 用法

```text
aitlog [flags] [keyword...]
aitlog search [flags] <keyword...>
aitlog sources [--json] [--backend auto|native|builtin]
aitlog version
```

无参数时，`aitlog` 打印根帮助并以 0 退出。查看或搜索需要旗标和（或）关键字（例如 `aitlog -n 50`）。

默认路由会查看或搜索日志。关键字对消息做不区分大小写的 **AND** 匹配。当关键字会与子命令名冲突时（例如 `version`），使用 `aitlog search …`。

用 `--expr` 做布尔匹配（`&` / `|` / `!` 或 `AND` / `OR` / `NOT`、括号、`"短语"`）。请给值加 shell 引号，避免 `|` 与 `!` 被 shell 解释。`--expr` 不能与位置关键字或 `-e` / `--regex` 同时使用。

监视式流式输出：`aitlog -f --expr '…'` 跟随新条目，只打印匹配项（命中前保持安静）。带 `-f` 时忽略 `-n` / `--head`；用 Ctrl+C 停止。

在 macOS 上，关键字 / 正则 / `--expr` 查询使用 `log show --predicate`，最多扩到最近一小时（stderr 会注明窗口）。要搜更宽的显式范围，请传 `--since`。

### 常用旗标

| 旗标 | 含义 |
|------|------|
| `-n, --lines` | 最近 N 条**匹配**记录（默认 1000；与 `-f` 同用时忽略） |
| `--head` | 取前 N 条匹配，而不是最近 N 条（与 `-f` 同用时忽略） |
| `-e, --regex` | 正则匹配 |
| `--expr` | 布尔关键字 / 短语表达式（`&\|!` 或 `AND/OR/NOT`） |
| `--case-sensitive` | 关键字 / 正则 / expr 区分大小写 |
| `--since` / `--until` | 时间窗口（`15m`、`2h`、`today`、RFC3339） |
| `-p, --priority` | 最低优先级（`emerg`……`debug`） |
| `-u, --unit` | 按 unit / service 过滤 |
| `-S, --source` | 逻辑源名，或文件路径 / glob |
| `-f, --follow` | 跟随新条目（关闭分页器；监视到 Ctrl+C） |
| `--json` | 向 stdout 输出 NDJSON |
| `--raw` | JSON 中包含原始行 |
| `--utc` | 时间戳用 UTC |
| `--no-pager` | 禁用交互分页器 |
| `--backend` | `auto`（默认）、`native` 或 `builtin` |
| `--quiet` | 抑制 stderr 上的提示 |
| `-V, --version` | 打印版本 |

在 TTY 上，长输出会打开类似 less 的分页器（`space` / `b` / `j` / `k` / `g` / `G` / `/` / `q`）。非 TTY 与 `--no-pager` 写普通行（每条匹配记录一行；消息中的嵌入换行展平成空格）。无匹配时以 0 退出，stdout 为空（除非 `--quiet`，否则 stderr 有提示）。跟随模式在命中前保持安静。

## 回退矩阵

过滤、搜索与分页都在 Go 中对归一化后的记录流执行。缺少原生工具只会减少能打开的存储。

| 平台 | 源 | 优先 | 内置回退 |
|------|----|------|----------|
| Linux | `system` | `journalctl -o json` | `/var/log/syslog`、`/var/log/messages`、`/dev/kmsg` |
| Linux | `kernel` | `journalctl -k` | `/dev/kmsg`、`/var/log/kern.log` |
| Linux | `auth` | `journalctl` | `/var/log/auth.log`、`/var/log/secure` |
| macOS | `system` | `log show --style ndjson`，然后 syslog 风格的 `log show` | `/var/log/system.log` |
| macOS | `install` | — | `/var/log/install.log` |
| Windows | `system` / `application` / `security` | Win32 `EvtQuery` | `wevtutil`，再往后是尚未实现的 Stage B `.evtx` |
| Windows | `setup` | — | `C:\Windows\Logs\…` |
| 任意 | `-S path` | — | 内置文件读取器（轮转与 `.gz`） |

`aitlog sources` 显示链路，并用 `*` 标出选中层。Windows 事件日志不支持 `-f`；文件源用 `-f --expr -S file`，或不用 follow 做历史 `--expr`。

## 示例

```bash
aitlog -n 50
aitlog error --since 1h
aitlog -e 'fail(ed|ure)' -n 100
aitlog -f --expr '(error|fail) & !debug'
aitlog --expr '(error OR fail) AND NOT timeout' -n 50 --since 1h
aitlog search --expr 'error|fail' --since 1h
aitlog --backend builtin -S /var/log/system.log -n 20
aitlog --json --no-pager -n 5
aitlog search version
aitlog sources --json
```
