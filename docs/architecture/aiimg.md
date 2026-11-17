# aiimg 内部实现

## 包分层

```text
cmd/img/main.go                → 精简 main；委托给 aiimg.Run
internal/img                     → 包 aiimg：Run、ExecuteWriters（路由）、WriteRootHelp
internal/img/registry          → 顶层 Register / AttachRoot / IsRegistered
internal/img/registry/all      → blank-import 各命令注册包
internal/img/extract           → 默认路由（不经 registry.Register）；包名 ≠ CLI 动词
internal/img/set               → 已注册的 `aiimg set`（JPEG EXIF 写入）
internal/img/clear             → 已注册的 `aiimg clear`（JPEG EXIF 清除）
internal/img/outpath           → set/clear 目标路径解析（AutoName、-o、确认）；不是命令
internal/img/version           → VERSION + ldflags + 已注册的 `aiimg version`
internal/pkg/toolname             → 各工具二进制名常量（见 `toolname` SSOT；含 `AITImg` 等）
internal/pkg/imgmeta              → Extract + Patch/PatchTo + Clear/ClearTo（JPEG 写入）+ 格式化器
internal/pkg/geocode              → Nominatim HTTP 反向地理编码
```

导入方向：`cmd` → `internal/img` → `extract` / `set` / `clear` / `outpath` / `registry` / `version` → `internal/pkg/*`。  
`internal/pkg` 不得导入 `internal/img`。  
`registry` 不得导入 `extract`。

### 路由

1. 空 argv，或单个 `-h` / `--help` / `help` → `WriteRootHelp`
2. `args[0]` 已注册（`version`、`set`、`clear` 等）→ 经 `AttachRoot` 走 Cobra
3. 否则 → `extract.Run`（旗标 + 文件），含 `-V` / `--version`

不支持 `help <cmd>`；用 `aiimg set --help`、`aiimg clear --help`、`aiimg version --help`。

### OCP：新增顶层命令

1. 创建 `internal/img/<cmd>/{register,cmd,help,run}.go`，在 `init` 中调用 `registry.Register`
2. 在 `registry/all/imports.go` 中 blank-import 该包
3. 不要手改 `cli_root.go` 中的 `AddCommand` 列表

与常见文件名冲突的命令名（例如 `version`）优先于默认 extract 路径。

### EXIF 写入路径

- `imgmeta.PatchTo` / `ClearTo` 使用 `dsoprea/go-exif/v3` + `dsoprea/go-jpeg-image-structure/v2`
- `Patch` / `Clear` 为原地写入封装（`dst == src`）
- 仅 JPEG（SOI 检查）；其他格式 → `ErrUnsupported`
- 经 `Dir(dst)` 下临时文件 + `replaceFile` 写入（Windows 安全重命名）
- CLI 默认：经 `outpath.Resolve` 写入兄弟文件 `*_aiimg.jpg`（TTY 覆盖提示；`-o` / `--inplace` / `-y`）
- 可写字段：Make、Model、Software、LensMake、LensModel、SerialNumber、DateTime、GPS、Altitude

## 文件布局

### `internal/img`

| 文件 / 目录 | 职责 |
|------------|------|
| `root.go` | `Run`（errhint；`errors.Is(imgmeta.ErrPartialFailure)`） |
| `cli_root.go` | `Execute` / `ExecuteWriters` / `newRootCmd`（仅 AttachRoot） |
| `help.go` | `WriteRootHelp`（extract 各段 + Commands） |
| `init_hooks.go` | `_ "…/registry/all"` |
| `registry/` | Register、AttachRoot、Entries、IsRegistered、RootHelpEntries、Deps |
| `extract/` | 默认路由 Cobra、`Run`、`WriteHelp`、地理编码测试钩子 |
| `set/` / `clear/` | 已注册的写入/清除子命令 |
| `outpath/` | AutoName、Resolve、confirm、输出哨兵（未注册） |
| `version/` | `Version()` / `Short()` + 已注册子命令 |

### `internal/pkg/imgmeta`

| 文件 | 职责 |
|------|------|
| `types.go` | `Info`、包注释 |
| `errors.go` | `ErrEmptyPath`、`ErrUnsupported`、`ErrPartialFailure`、`ErrNothingToChange` |
| `extract.go` / `extract_helpers.go` | `Extract` + 尺寸/格式辅助 |
| `patch.go` / `patch_jpeg.go` | `PatchFields`、`ClearMask`、`Patch`/`PatchTo`、`Clear`/`ClearTo` |
| `format_human.go` / `format_helpers.go` | 锁定摘要字段 |
| `format_json.go` | JSON 数组（+ 可选嵌套 `"exif"`） |
| `format_full.go` / helpers | `--all` Field/Value 表 |
| `format_full_structured.go` | `--all` 内结构化字符串展开 |

### `internal/pkg/geocode`

| 文件 | 职责 |
|------|------|
| `nominatim.go` | `Client.Reverse`（Nominatim） |

## 数据流

1. `Extract(path)` 打开文件，使用 `evanoberholster/imagemeta`（尽力解码）。
2. 成功/部分解码保留 `Info.Exif`（`*exif.Exif`），供 `--all` 使用。
3. 若存在 GPS 且未设置 `--no-geocode`，CLI 调用 `geocode.Client.Reverse`。
4. 格式化器打印锁定的 `Info` 字段；`--all` 额外通过 `FormatFullHuman` / 嵌套 JSON `"exif"`，来自对 `Info.Exif` 的库序列化。
5. `set`/`clear` 校验旗标 → `outpath.Resolve` → `imgmeta.PatchTo`/`ClearTo`。

## 验收

**IM\*** / **IM-FULL\*** / **IM-TABLE\***：提取与 `--all` 格式化（不变）。

**SC1–SC10**：set/clear JPEG 写入（`TestPatch_*`、`TestClear_*`、`TestSetCmd_NoFlags`、`TestClearCmd_NoFlags`、`TestCLI_SetAndClearMake` / `TestCLI_SetPartialFailure` 配合 `--inplace`、根帮助 Commands）。

**OW1–OW14**：默认兄弟文件写入 + 覆盖提示（`TestCLI_SetDefaultNewFile_NonTTY`、`TestOutpath_Confirm*`、`TestCLI_SetOutput*`、`TestAutoName_Collision`、`TestCLI_ClearDefaultNewFile_NonTTY`、`TestPatchTo_WritesToDst`、输出旗标相关 help 测试、`TestCLI_SetOutputEmpty`）。
