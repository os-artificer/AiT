# aitvid

> **命名：** 二进制 `aitvid`，插件命令 `vid`（`ait vid …`），Go 包 `aivid`。

从视频文件或目录中检索文字（OCR）与物体（可选视觉模型），并把命中关键帧落盘（JPEG + 侧车 JSON + 可选像素烧录）。

## 构建

```bash
make build
./bin/aitvid --help
```

## 依赖

硬依赖：`ffmpeg` 与 `ffprobe`（PATH）。OCR：`tesseract`（`--local-only` / `--mode text` 时为硬依赖）。物体检索另需配置 `ai.vision.model` 与可用的多模态 API。

```bash
aitvid deps
aitvid deps --install          # TTY 确认
aitvid deps --install --yes    # 非 TTY
```

`CGO_ENABLED=0`：不内嵌 ffmpeg/tesseract。

## 用法

```bash
aitvid search ./clip.mp4 --query "京A" --mode text --local-only
aitvid search ./clips -r --query bike --camera-id cam03
aitvid search ./clip.mp4 --query plate --json -o ./hits
aitvid frames ./clip.mp4 -o ./aitvid-frames
aitvid version
```

常用旗标：`--mode text|object|auto`（默认 `auto`：先 OCR，无命中再 VLM）、`--local-only`、`--fps`（默认 0.5）、`--max-frames`（默认 30）、`--max-videos`（默认 **0 = 不限数量**；显式 `N>0` 时超出报错）、`--timeout`（默认 10m；大目录串行处理时请加大，例如 `--timeout 2h`）、`--save-hits` / `--no-save-hits`、`--annotate` / `--no-annotate`、`-o/--out-dir`、`--camera-id`、`--json`。

命中落盘目录：`<out-dir>/<camera_id|unknown>/<stem>_t<t_sec>s_<id>.jpg` 与同名 `.json`。

## 隐私

`--mode object` / `auto` 在 OCR 未命中时可能把抽帧上传到配置的云端或本地 vision endpoint。用 `--local-only` 关闭上传。车牌/电话等为 OCR+正则尽力而为，不宣称专用识别精度。

## Agent 风险

| 命令 | 级别 |
|------|------|
| `vid frames` / `vid deps` / `vid search --no-save-hits` | Read（可自动放行） |
| `vid search`（默认落盘） / `vid deps --install` | Low（需确认） |

## 另见

- 架构：[../architecture/aitvid.md](../architecture/aitvid.md)
- 配置：`ait agent config set ai.vision.model <model>`
