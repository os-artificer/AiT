# aitkeep

按 `keepalive.yaml` 监管进程，并用 `pong` 响应 `GET /ping`。

发布的二进制是 `aitkeep`。工具箱插件命令是 `keep`（`ait keep …`）。

## 构建

```bash
make build   # → bin/aitkeep
./bin/aitkeep --help
```

## 用法

```text
aitkeep [command]
aitkeep serve [--listen addr]
aitkeep status
aitkeep start NAME
aitkeep stop NAME
aitkeep reload
aitkeep add --name NAME [--cwd DIR] [--cpu P] [--mem BYTES] [--disk-bytes N] [--disk-percent P] [--read-bps N] [--write-bps N] [--io-percent P] -- COMMAND...
aitkeep remove NAME
aitkeep version
```

无参数时打印帮助。`serve` 是守护进程。`version` / `-V` 在进入 Cobra 前处理。

`status`、`reload`、`start`、`stop` 经管理端点联系正在运行的 `serve`。守护进程未起时返回错误并带 `Try:`（提示先 `serve`）。

`add` / `remove` 读写 `keepalive.yaml`。即使 `serve` 未运行，`remove` 仍会删 yaml 条目；`add` 可写 yaml，需已运行的 `serve`（或之后 `reload`）才开始监管新 unit。

配置、单实例锁与管理端点位于 AppRoot（`AIT_HOME`，否则为平台数据目录）下：

| 文件 | 作用 |
|------|------|
| `keepalive.yaml` | 监听地址与 unit（模式 0600） |
| `keepalive.lock` | 每个 AppRoot 只允许一个 `serve` |
| `keepalive.sock` | Unix 管理套接字 |

在 Windows 上，管理端点是由 AppRoot 路径派生的命名管道，因此两个家目录不会共用同一根管道。

`GET /ping` 返回 `200`、`text/plain`，正文为 `pong\n`。其他方法或路径返回 `404`。默认监听地址是 `127.0.0.1:8757`（`--listen` 或 yaml 的 `listen`）。写进 yaml 的监听地址在下一次启动 `serve` 时生效。

## Units

```yaml
listen: 127.0.0.1:8757
units:
  - name: demo
    command: ["sleep", "3600"]
    cwd: /var/lib/demo
    limits:
      cpu_percent: 20
      memory_bytes: 268435456
      disk_space_bytes: 1073741824
      disk_space_percent: 90
      disk_read_bps: 1048576
      disk_write_bps: 1048576
      disk_io_percent: 80
```

限制为 0 表示关闭该维度。绝对上限与百分比同时配置时，任一超限都会触发该维度的动作。

| 限制 | 含义 | 超限时 |
|------|------|--------|
| `cpu_percent` | 相对一核的百分比（`150` 表示 1.5 核） | 内核接受时做硬帽；否则在每个 1s 采样内暂停，持续超限运行 10s 后杀死 |
| `memory_bytes` | 进程树 RSS | Linux 的 `memory.max` 或 Windows job 内存限制；否则墙钟 10s 后杀死。进程不会被暂停 |
| `disk_read_bps` / `disk_write_bps` | 进程树每秒字节数 | 可用时写 Linux `io.max`；否则与 CPU 相同的暂停后杀死路径 |
| `disk_io_percent` | 块设备忙碌百分比，含其他进程 | 仅当本进程在窗口内也有读写时，10s 后杀死。空闲进程不动。无设备统计 → `io=unsupported` |
| `disk_space_bytes` | `cwd` 下普通文件字节总和 | 终止且不重启 |
| `disk_space_percent` | 容纳 `cwd` 的文件系统已用块百分比 | 终止且不重启 |

空的 `cwd` 会跳过磁盘空间百分比与设备忙碌百分比。目录遍历触达条目预算，或文件系统 stat 失败，对已配置的磁盘空间限制视为超限。

异常退出会按 1s～30s 退避重启。一次运行保持 30s 会重置退避与重启计数。10 分钟内异常重启 5 次会把 unit 标为 `failed`。退出码 0 不重启。磁盘空间失败后不会重启，直到之后的 `start` 看到用量回到限制以下。

`reload` 需要 `serve` 在跑；只重启规格有变化的 unit。用 `stop` 停掉的 unit，在规格未变时，`reload` 后仍保持停止。`add` 会替换同名 unit。`remove` 即使 `serve` 未运行也会删除 yaml 条目。

`status` 需要 `serve`；打印每个 unit 的状态，以及对每个非零限制的当前采样。

## 示例

```bash
aitkeep serve
aitkeep add --name demo --cpu 20 -- sleep 3600
aitkeep status
aitkeep stop demo
aitkeep remove demo
curl -fsS http://127.0.0.1:8757/ping
```
