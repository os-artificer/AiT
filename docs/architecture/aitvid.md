# aitvid 内部实现

## 包分层

```text
cmd/vid/main.go                 → 精简 main；委托 aivid.Run
internal/vid                    → 包 aivid：Run、ExecuteWriters、WriteRootHelp
internal/vid/registry           → 顶层 Register / AttachRoot
internal/vid/registry/all       → blank-import search/frames/deps/version
internal/vid/search             → aitvid search（OCR + 可选 VLM + 落盘）
internal/vid/frames             → aitvid frames（仅抽帧）
internal/vid/deps               → aitvid deps / deps --install
internal/vid/version            → VERSION + ldflags
internal/pkg/vidsearch          → RunSearch 管道
internal/pkg/vidsearch/hit      → Hit 叶子类型（打破 import cycle）
internal/pkg/vidsearch/deps     → Check / Install（假 Runner）
internal/pkg/vidsearch/discover → ListVideos
internal/pkg/vidsearch/frames   → ffmpeg/ffprobe 抽帧
internal/pkg/vidsearch/ocr      → tesseract 索引与匹配
internal/pkg/vidsearch/export   → SaveHitFrame（annotate + sidecar + EXIF）
internal/pkg/vidsearch/format   → 人类 / NDJSON
internal/pkg/vidsearch/vision   → OpenAI-compatible 多模态 Analyze
```

导入方向：`cmd` → `internal/vid` → `vidsearch/*` → 其它 `internal/pkg`。`internal/pkg` 不得导入 `internal/vid`。`internal/vid` 不得导入 `internal/ssh`。

## 检索管道

1. `deps.Check`（缺硬依赖 → errhint + Try `deps --install`）
2. `discover.ListVideos`（文件点名或目录；默认不限数量；仅当显式 `--max-videos N>0` 时超限报错）
3. 串行每视频：`frames.Extract` → OCR（非 object）→ 必要时 `vision.Analyze`
4. `--save-hits`：`export.SaveHitFrame`
5. `format.WriteHuman` / `WriteNDJSON`（命中 stdout；错误 stderr）

`auto`：OCR 有命中则不调 VLM；`object` 且无 vision → `ErrVisionRequired`。

## 配置

- `ai.vision.model`：发图前校验；缺省不盲目用 chat model
- 走现有 `ai.experimental` / provider keyring

## 验收

见 [change-acceptance.mdc](../../.cursor/rules/change-acceptance.mdc) **VD\***（P0 VD1–VD10 + VD17*；P1 VD11–VD16）。
