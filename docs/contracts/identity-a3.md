# A3 / A6 身份状态存放

本文只记录 A3（建档续答）和 A6（区号 / 语言偏好）的存放方式。产品口径以同目录《新用户流程 A1 身份契约（2026-11-25）》v0.1.0-C 为准。没有新的身份路径，请求和响应形状不改。

## 重启后还在

PostgreSQL 保存：

- 学习者账号，包括稳定 `user_id` 和已绑定的登录号码（`phone_region` + `phone_national`）
- 建档 7 步已写入的答案。未写入的步保持空。重启后从第一道未完成步续答，不能跳步
- `locale`（`zh-Hans` / `zh-Hant`）和偏好 `region_code`（`+86` / `+852`）
- 会话的访问令牌摘要、刷新令牌摘要和到期时间。摘要是 SHA-256，不保存令牌原文

`PUT /v1/identity/profile` 只更新偏好。绑定登录号码不变。用偏好里的新区号去登录，找不到这个账号。

`has_first_pack` 存在 `identity_accounts`。身份 HTTP 不写它。知识域在首次导入作业进入 `map_ready`，或 `POST /v1/knowledge/first-pack/skip` 成功时，把它置为 true。测试夹具仍可以直接设置。跳过记录在知识域表，见 `knowledge-a4.md`。

## 重启后不在

验证码摘要、发码 / 验码限流、幂等记录仍在 API 进程内存。短信仍是进程内 mock，不接供应商。验证码明文不入库、不写日志。

## 下一步

`next_step` 沿用 A2：建档未完成是 `NEED_ONBOARDING`，建档完成且没有首包是 `NEED_FIRST_PACK`，两者都完成是 `READY_HOME`。这是方案 §9 H3 的实施假设，不是已确认产品口径。

表在 API 启动时创建（`identity_accounts`、`identity_onboarding_answers`、`identity_sessions`）。本地库使用 Compose 里的 PostgreSQL，宿主机端口 `4713`。
