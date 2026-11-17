# aiimg

> **命名：** 二进制 `aitimg`，插件命令 `img`（`ait img …`），Go 包 `aiimg`。

提取常见格式的 EXIF（设备、拍摄时间、GPS、逆地理地址）。也可对 JPEG **设置** / **清除** 可写 EXIF（默认写同目录兄弟文件）。

## 构建

```bash
make build
./bin/aitimg --help
```

## 用法

```bash
aitimg photo.jpg
aitimg --json photo.jpg shot.png
aitimg --all photo.jpg              # 锁定摘要 + 完整结构化 EXIF
aitimg --json --all photo.jpg       # 锁定键 + 嵌套 "exif" 对象
aitimg --no-geocode photo.jpg       # 跳过网络；Location 显示 0；有 GPS 时仍显示
aitimg --version                    # 或 -V；也可：aitimg version
aitimg version --help               # 子命令帮助（不是：aitimg help version）

aitimg set --make Canon photo.jpg                 # 默认写入 photo_aiimg.jpg
aitimg set --make Canon --inplace photo.jpg       # 覆盖原文件
aitimg set --gps 31.2304,121.4737 -o out.jpg photo.jpg
aitimg set --datetime "2020:06:07 01:30:00" photo.jpg
aitimg clear --gps --make photo.jpg
aitimg clear --all --inplace photo.jpg
```

无文件参数时，`aitimg` 打印根帮助并以退出码 0 退出。

`aitimg version` 打印构建版本（优先于名为 `version` 的文件）。

### set / clear（仅 JPEG）

可写 / 可清除的锁定字段：

- Make、Model、Software、LensMake、LensModel、SerialNumber
- DateTime
- GPS（纬度/经度）、Altitude

不可写：File、Format、Width、Height、Location（Address 仅来自反向地理编码）。

`set` 至少需要一个字段旗标。`clear` 至少需要一个字段旗标或 `--all`。

**写入目标（相对早期默认原地写入为 BREAKING）：**

- 默认（非 TTY）：写入同目录兄弟文件 `*_aiimg.jpg`（冲突时为 `*_aiimg_N.jpg`）；源文件不变
- 默认（TTY）：在 stderr 上提示一次 `Overwrite original file(s)? [y/N]`（`y` = 覆盖；其他输入 = 兄弟文件）
- `-o` / `--output path`：写入该路径（仅单输入；可覆盖已有文件；不提示）
- `--inplace` / `-y`：覆盖每个输入，不提示
- `-o` 不能与多个输入或与 `--inplace`/`-y` 合用
- `-o` 拒绝已存在的目录

PNG/TIFF/WebP 返回不支持。

### 输出

人类可读与 JSON 模式始终打印每一个锁定字段。缺失值显示为 `0`（GPS 为 `0.000000, 0.000000`；Altitude 为 `0 m` / JSON `0`）：

- File、Format、Width、Height
- Make、Model、Software、LensMake、LensModel、SerialNumber
- DateTime
- Location（地理编码成功时为 Nominatim 地址；否则为 `0`）
- GPS（EXIF 坐标；缺失时为零）
- Altitude（有则显示，否则为 `0`）

JSON 模式始终输出 **JSON 数组**（每个成功文件一个对象），键与锁定字段相同。

#### `--all`（结构化 EXIF）

在锁定摘要之后（或在 `--json` 下的 `"exif"` 下），打印由 `imagemeta` 解析的 TIFF/EXIF IFD 树：

- 人类可读模式：每个区段为两列 **Field / Value** 表（列对齐）
- 根级：ImageType、CameraSerial、CameraMakeID（存在时）
- 区段顺序：**IFD0 → ExifIFD → GPS → IFD1 → IFD2 → MakerNote**（省略空/空值区段）
- 存在时包含曝光/光学字段（ExposureTime、FNumber、ISO、FocalLength 等）
- 完整转储跳过缺失/空值（与锁定摘要始终显示 `0` 的策略不同）
- **字符串字段**（例如 `UserComment`）：自动检测 **JSON**、**XML** 或 **YAML**。结构化文档展平为点分路径（`UserComment.AIGC.Label`）；否则显示 **原始** 字符串。嵌套结构化字符串最多展开到深度 3。长值换行到续行（从不截断）
- **不是** 原始 APP1 标签转储；XMP/ICC/IPTC 及未映射的私有标签不在范围内
- `--json --all` 仍输出嵌套 `"exif"` 机器可读 JSON（形状不变）

### 网络

默认在存在 GPS 时，`aitimg` 调用 OpenStreetMap [Nominatim](https://nominatim.org/) 反向地理编码 API（`accept-language=zh-CN,en`）。需要网络访问，并遵循 Nominatim 使用策略（设置 User-Agent；请求间隔 ≥1s）。

离线使用请加 `--no-geocode`。

### 格式

**读取**：JPEG、PNG；TIFF/WebP 视解码器能力。**set/clear**：仅 JPEG。
