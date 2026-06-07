# 更新日志

**语言：** [English](CHANGELOG.md) · 简体中文（本页）

本文仅记录 **产品与 CLI** 的功能性变更，不包含纯文档类更新。
发版后请将 `[未发布]` 中的条目移至带日期的版本段。

## [未发布]

- 【优化】会话连接唯一键扩展为 username+host+port+alias；同一目标机可用多个 `--alias` 保存独立会话；无 flag login 仅更新无 alias 匹配行。

## v1.0.0-beta.1

- 【新增】AI 增强 SSH 客户端，可以在SSH会话内通过 !! '问题' 、!ait ai ask 、!ait ai agent 或!ait ai激活AI能力。
- 【新增】自带文件传输功能，在SSH会话内通过!ait get/put 实现文件下载与上传。
- 【新增】在SSH会话内随时进行local与remote模式间的切换。
