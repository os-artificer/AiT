# 未发布

### aitvid

- 【Added】官方插件 `aitvid`（`vid`）：目录/文件视频 OCR 检索、可选 VLM、`deps --install`、命中关键帧落盘（JPEG + 侧车 + 烧录）。
- 【Changed】`search`/`frames` 目录发现默认不限视频数量（`--max-videos` 默认 0）；仅显式 `N>0` 时超限报错。大目录请酌情加大 `--timeout`。

### aitkeep

- 【Added】`aitkeep serve` 按 `keepalive.yaml` 监管进程，`GET /ping` 的响应正文是 `pong`。

### aihost

- 【Added】`aithost net --conn` 列出连接。表上有 PROTO、STATE、LOCAL、REMOTE、PID、PROCESS、AGE 和 CREATED。
