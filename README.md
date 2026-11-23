# AiT

**AiT** 是面向运维的跨平台（Linux / macOS / Windows）命令行工具箱。当前覆盖终端与 SSH 工作流，并随版本继续增加工具。

**工具：**

- [`ait`](cmd/ait) — Agent 交互入口（默认进入 Loop；`ait agent` 管理命令；`ait shell` 为工具箱 REPL）。说明见 [ait](docs/guide/ait.md)。
- [`aitssh`](cmd/ssh) — SSH 客户端（产品文档标题 [aissh](docs/guide/aissh.md)；插件命令 `ssh`；Go 包名 `aissh`；不含 AI）。
- [`aitimg`](cmd/img) — 图像元数据提取（EXIF、GPS、逆地理编码地址）。说明见 [aiimg](docs/guide/aiimg.md)（插件 `img`）。
- [`aitlog`](cmd/log) — 系统日志查看与搜索（Linux / macOS / Windows）。说明见 [ailog](docs/guide/ailog.md)（插件 `log`）。
- [`aithost`](cmd/host) — 主机 CPU / 内存 / 磁盘 / 网络 / 进程监视。说明见 [aihost](docs/guide/aihost.md)（插件 `host`）。
- [`aitkeep`](cmd/keep) — 进程监管（插件命令 `keep`；`GET /ping` 返回 `pong`）。说明见 [aitkeep](docs/guide/aitkeep.md)。
- [`aitvid`](cmd/vid) — 视频内容检索（OCR / 可选 VLM；插件 `vid`）。说明见 [aitvid](docs/guide/aitvid.md)。

## 能力概览

- 兼容 OpenSSH 的连接方式（`-J`、config Host 别名、经 `ssh -G` 解析）
- 会话书签：`--alias`、`session --load`、`session --switch`
- 会话内 `!ait` 元命令（挂起、链式跳转、传输）
- 默认由 AiT 管理远程提示符（`AiT user@host:…>`）；设置 `AIT_REMOTE_SYSTEM_PROMPT=1` 可保留远端原生提示符
- 运维 Agent 经 **`ait`** / **`ait agent`** 进入（Agent Loop、connect、model、provider、ollama、config）；`aitssh` 不含 AI 子树
- 当前 SSH 会话上的 `!ait get` / `!ait put`，以及独立的 `aitssh get` / `put`（SFTP）
- 可选 `--tmux`，断开后仍保持远端 shell
- 本地 `aitssh shell` REPL，支持历史与 UTF-8 行编辑

## 示例

```bash
# 1）给主机起别名，之后可恢复
aitssh deploy@prod --alias prod
aitssh session --list
aitssh session --load prod

# 2）经跳板机连接（与 ssh -J 相同；Host 别名经 ssh -G 解析）
aitssh -J bastion app.internal

# 3）经 ait 进入 Agent Loop 或远程连接
ait agent config init
ait agent model list
ait                              # 本机 Agent Loop
ait agent connect user@host      # 远程 Agent

# 4）在当前会话中传输（单参数时本地目标默认为 ./basename）
!ait get ./build/app-linux-amd64
!ait put ./patch.tar.gz

# 5）tmux 保活：断开后远端工作仍在
aitssh devhost --tmux
# 之后：aitssh session --load devhost --tmux

# 6）交互式 Agent Loop（默认开启；仅在关闭后需要再打开）
ait agent "summarize this log pattern: connection reset"
```

## 安装

**前置条件：** 从源码构建需要 Go **1.25+**；`aitssh` 的 connect / `load` 解析 `~/.ssh/config` 时，PATH 上需要 **OpenSSH 客户端**（`ssh`）。

### 安装 `ait`

```bash
go install github.com/os-artificer/AiT/cmd/ait@latest
ait --help
```

请确认 `$(go env GOPATH)/bin` 已加入 PATH。

### 安装官方插件（`ssh`、`log`、`img`、`host`、`keep`）

有可用的 [GitHub Release](https://github.com/os-artificer/AiT/releases) 后，用 `ait shell` 打开工具箱 REPL。无参的 `ait` 会进入 Agent Loop；未配置 AI 时会退出并打印未配置提示：

```bash
ait shell
```

```text
ait> install              # 全部官方插件
ait> install ssh keep     # 按插件命令名安装一个或多个
ait> update               # 更新已安装的插件
ait> upgrade              # 更新 ait 与已安装的插件
ait> help install
```

遗留安装名（例如 `aissh` → `ssh`）仍可解析。二进制放在正在运行的 `ait` 可执行文件旁（布局与 `make install` 相同）。细节见 [docs/guide/ait.md](docs/guide/ait.md#install--update--upgrade)。

### 从源码构建（开发）

```bash
git clone https://github.com/os-artificer/AiT.git
cd AiT
make build
./bin/ait --help
```

本地开发时，`make build` 还会产出 `bin/aitssh`、`bin/aitimg`、`bin/aitlog`、`bin/aithost`、`bin/aitkeep`、`bin/aitvid`。若要在没有 Release 的情况下把插件装进 `bin/`，运行 `make install`。有 Release 之后，先运行 `ait shell`，再在 `ait>` 提示符下执行 `install`。见 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 文档

| 文档 | 路径 |
|------|------|
| 项目结构 | [docs/architecture/overview.md](docs/architecture/overview.md) |
| aissh 实现 | [docs/architecture/aissh.md](docs/architecture/aissh.md) |
| aissh 使用说明 | [docs/guide/aissh.md](docs/guide/aissh.md) |
| ait 实现 | [docs/architecture/ait.md](docs/architecture/ait.md) |
| ait 使用说明 | [docs/guide/ait.md](docs/guide/ait.md) |
| aiimg 实现 | [docs/architecture/aiimg.md](docs/architecture/aiimg.md) |
| aiimg 使用说明 | [docs/guide/aiimg.md](docs/guide/aiimg.md) |
| ailog 实现 | [docs/architecture/ailog.md](docs/architecture/ailog.md) |
| ailog 使用说明 | [docs/guide/ailog.md](docs/guide/ailog.md) |
| aihost 实现 | [docs/architecture/aihost.md](docs/architecture/aihost.md) |
| aihost 使用说明 | [docs/guide/aihost.md](docs/guide/aihost.md) |
| aitkeep 实现 | [docs/architecture/aitkeep.md](docs/architecture/aitkeep.md) |
| aitkeep 使用说明 | [docs/guide/aitkeep.md](docs/guide/aitkeep.md) |
| aitvid 实现 | [docs/architecture/aitvid.md](docs/architecture/aitvid.md) |
| aitvid 使用说明 | [docs/guide/aitvid.md](docs/guide/aitvid.md) |
| Python E2E 测试 | [tests/README.md](tests/README.md) |
| 仓库内 SSH/HTTP mock | [docs/architecture/testenv-mock.md](docs/architecture/testenv-mock.md) |
| 贡献指南 | [CONTRIBUTING.md](CONTRIBUTING.md) |
| 代码布局 | [AGENTS.md](AGENTS.md) |
| 行为准则 | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| 变更记录 | [CHANGELOG.md](CHANGELOG.md) |
