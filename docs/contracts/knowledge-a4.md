# A4 实现说明

本文只记录 A4 的实现选择。产品口径以同目录《新用户流程 A4 知识包与导入受理契约 v0.1.1-C（2026-11-26）》和《锁定说明 · 新用户流程 A4 契约 v0.1.1-C（2026-11-26）》为准。

H3 仍是实施假设，不是已确认产品口径。A4-O2（预签名秒数、材料字节上限、查询轮询间隔、幂等键保留时长）和 A4-O3（失败 `failure_code` 清单）仍开放。下面的数字是当前进程默认值，不是契约锁定项。本进程不枚举挖掘或出题失败码。查询轮询间隔不由服务端规定。

## 写入

- 创建空知识包、申请预签名、导入 `202` 都不把 `has_first_pack` 从 false 改成 true。
- `has_first_pack` 置真只有两处：跳过接口成功，或导入作业进入 `map_ready`。HTTP 受理把作业写成 `accepted` 并入队，不把它写成 `map_ready`。`map_ready` 只出现在测试夹具里，留给以后的 A5。
- 跳过不创建知识包，也不创建导入作业。建档未完成时 `next_step` 仍是 `NEED_ONBOARDING`。建档已完成时是 `READY_HOME`。
- 已经 `has_first_pack=true` 再跳过：返回 200，标志保持 true，不删除已有知识包或作业。
- `has_first_pack` 写在身份表。跳过另写 `knowledge_first_pack_skips`。两处都在 PostgreSQL，新进程能读回。

## 进程默认

| 项 | 当前进程默认 |
|---|---|
| 预签名有效秒数 | 900 |
| 材料字节上限 | 33554432 |
| 幂等记录保留 | 24 小时 |
| 幂等键 | 与身份域相同，1–200 个可见 ASCII |

命中字节上限时，响应 `details.max_bytes` 带上当前上限。这不是已锁产品数字。

## 存放

PostgreSQL 保存知识包、材料元数据（含服务端生成的对象键和可选展示文件名）、导入作业、跳过记录、导入与跳过的幂等键。材料字节不进库，也不进 API 响应。对象键不作为单独字段返回。文件名不参与对象键。

导入成功后向 `QLUTCH_WORKER_QUEUE` 放入只含 `job_id` 的消息。现有工人进程仍把载荷当不透明字节收下，不改作业状态。

同一学习者、同一幂等键、同一请求体重放时，导入返回当前作业，而不是受理当时的快照。因此作业后来进入 `map_ready` 后，重放仍能看到。

## 环境变量

名称沿用 `.env.example` 里已有的 `QLUTCH_S3_ENDPOINT`、`QLUTCH_S3_REGION`、`QLUTCH_S3_BUCKET`、`QLUTCH_S3_ACCESS_KEY`、`QLUTCH_S3_SECRET_KEY`、`QLUTCH_WORKER_QUEUE`、`QLUTCH_WORKER_HEARTBEAT`。本文不写值。云上的值部署时再填，不进仓库。对象存储只走 S3 兼容签名，领域代码不出现云厂商 SDK。
