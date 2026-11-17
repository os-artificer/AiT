# aitkeep 实现说明

## 包分层

```text
cmd/keep/main.go
  → internal/pkg/keepalive（包名 keepalive：Run、ExecuteWriters、Serve、Supervisor）
       → platform.DataDir（AppRoot）
       → helpdoc / errhint / buildinfo / toolname
```

无 `internal/keep` 应用包。守护进程与 CLI 共用 `keepalive`（ping、监管、unit 文件）。`cmd/keep` 只调 `keepalive.Run`。

导入方向：`cmd/keep` → `internal/pkg/keepalive` → 其他 `internal/pkg/*`。`keepalive` 不得导入 `internal/ssh`、`internal/ait`、`internal/host` 或 `internal/log`。

插件命令名是 `keep`（`toolname.Keep`）。二进制名是 `aitkeep`（`toolname.AITKeep`）。官方发现使用内嵌 manifest：`internal/pkg/aitplugin/manifests/aitkeep/plugin.json`。

## 路由

| argv | 路由 |
|------|------|
| 空 / `-h` / `--help` / `help` | `WriteRootHelp` |
| `-V` / `--version` / `version` | `Version()`（进入 Cobra 命令树之前） |
| `ait-plugin-manifest` | 内嵌 manifest JSON |
| `add`、`reload`、`remove`、`serve`、`start`、`status`、`stop` | 来自 `commandOrder` 的 Cobra 命令 |
| 其他 | errhint + Try |

## 控制面

`Serve` 接受父 context，获取 `keepalive.lock`，监听 `GET /ping`，并打开管理端点（`keepalive.sock`，或由 AppRoot 哈希命名的 Windows 管道）。`reload` 与 `start` 使用该父 context。

每个 unit 槽位有一代际。`reconcile` 在 `reflect.DeepEqual` 判定规格未变时，让正在运行的 unit 继续跑，让已停止的 unit 保持停止；停止已消失的运行中 unit；重启规格已变的 unit。`Watch` 协程仅在代际仍为当前时写状态。

## 限制

采样汇总进程树（CPU、RSS、IO 字节）。Linux 在 `exec` 前把子进程加入临时 cgroup，并在 `runProcess` 返回时删除该目录。硬帽在 Linux 上是 `cpu.max`、`memory.max`、`io.max`；在 Windows 上是 Job Object CPU 速率与 job 内存。`memory.high` 不是硬帽。当 `cpu_percent / NumCPU` 超过整机 100% 时，Windows CPU 速率会被拒绝。

`Decide` 不共用一把 sustain 时钟：

- CPU 与进程字节 / 秒：在 `SampleEvery` 内暂停；暂停时间不计入 duty 时钟。
- 内存与设备 `%util`：不暂停；时钟是墙钟。设备 `%util` 仅在本进程于该窗口有读写字节时才杀死。
- 磁盘空间字节或文件系统百分比：终止并返回 `ErrDiskFailed`（不重启）。触达 `DiskWalkBudget` 的遍历视为同样失败。
