# Qlutch

Qlutch 的单仓空骨架。这里有三端壳、后管壳、API、作业进程和共享契约，没有业务功能。

## 本地地址

开发端口放在 `4710`–`4716`。这组端口避开 `3000`、`5173`、`5432`、`6379`、`8080`、`9000`，方便和别的本地项目一起开。

| 用途 | 地址 |
|---|---|
| API 根路径 | `http://127.0.0.1:4710` |
| 存活探针 `healthz` | `http://127.0.0.1:4710/healthz` |
| 就绪探针 `readyz` | `http://127.0.0.1:4710/readyz` |
| 后管 | `http://127.0.0.1:4712` |
| PostgreSQL | `127.0.0.1:4713` |
| Redis | `127.0.0.1:4714` |
| S3 兼容存储 | `http://127.0.0.1:4715` |
| 对象存储控制台 | `http://127.0.0.1:4716` |

作业进程不监听端口。`healthz` 只表示 API 进程还活着。`readyz` 才检查数据库、Redis 和对象存储是否可达。

Android 模拟器访问宿主机时用 `http://10.0.2.2:4710`，默认构建就是这个地址。iOS 模拟器和鸿蒙默认产品用 `http://127.0.0.1:4710`。E1 联调另选 Android 参数 `-Pqlutch.e1.api=true`、iOS scheme `Qlutch-E1`、鸿蒙 product `e1`，根地址是 `http://47.93.180.37`（HTTP 80）。

## 启动

需要 Go、Node.js 当前 LTS、pnpm、[Task](https://taskfile.dev) 和 Docker。具体版本以 `go.mod`、`package.json` 的 `engines`、`pnpm-lock.yaml`、`apps/android/gradle/libs.versions.toml` 和 `infra/local/compose.yaml` 为准。落地时钉住的是 Go 1.27.4、Node.js 26.13.1、pnpm 12.21.0。

在仓库根目录：

```bash
cp .env.example .env
task bootstrap
task up
task gen
```

然后分三个终端：

```bash
task api
task worker
task admin
```

`.env` 不提交。样例里的数据库和对象存储口令只给本机 Compose 使用。

三端用各自的官方 IDE 打开：

- iOS：`apps/ios/Qlutch.xcodeproj`
- Android：`apps/android`
- 鸿蒙：`apps/harmony`

底栏只有五项：首页、做题、PK、好友、我的。五项都是占位，没有导入、地图、练习或社交。

## 常用任务

| 任务 | 作用 |
|---|---|
| `task bootstrap` | 安装依赖，并准备 golangci-lint |
| `task up` / `task down` | 启动或停止本地 PostgreSQL、Redis、S3 兼容存储 |
| `task gen` | 按 OpenAPI 生成 Go、TypeScript、Swift、Kotlin、ArkTS 客户端 |
| `task lint` | 平台 lint、后管检查、三端结构检查 |
| `task test` | `go test`、契约模型测试、后管单元测试 |
| `task api` / `task worker` / `task admin` | 分进程启动 |
| `task android` | 起模拟器、编译安装、看启动页 |
| `task android-welcome` | 同上去掉启动动画，直接看欢迎页 |
| `task android-cold` | 冷启动模拟器，并先跑 Android 单元测试 |

`packages/contracts` 变更后必须跑 `task gen`，并且生成物不能有未提交的差异。生成文件标了 `DO NOT EDIT`。

## 依赖方向

客户端和后管只依赖契约生成的客户端，不依赖服务端内部包。`api` 和 `worker` 是两个进程，都引用 `services/airuntime`。`airuntime` 不依赖页面或后管。模型调用只走它的 `Call`，未配置提供方时返回「未配置提供方」。三端保持各自的官方工程，不进 pnpm workspace，也不进 `go.work`。

学习者身份走短信验证码、会话和建档接口。账号、7 步建档、语言 / 区号偏好、会话令牌摘要和 `has_first_pack` 在本地 PostgreSQL。身份表或知识表创建失败时 API 进程退出，不开始监听。验证码、限流和身份幂等记录仍在 API 进程内存。短信是 mock，不接真实供应商。`QLUTCH_ENV=staging` 且 `QLUTCH_STAGING_OTP_TOKEN` 非空时，API 才注册内测取码接口；其他环境不回传验证码、不记明文。作业进程在 `QLUTCH_ENV=staging` 时，把整段正文正好是「无法完成挖掘」或「材料无法读取」的文本材料记成等待失败，其他材料仍走原来的挖掘。知识包、材料元数据、导入作业、跳过记录和知识域幂等键也在 PostgreSQL。材料字节走 S3 兼容预签名，不经过 API。导入受理后作业停在 `accepted` 并入队，API 不把作业改成 `map_ready`。工人进程再经 `airuntime` 跑 S01，地图落库后才把作业改成 `map_ready`，并记 S01 用量。内部域中间件仍是未接入的占位。没有微信或支付宝登录。本地对象存储走 S3 兼容请求签名，代码里没有云厂商 SDK。

## 还没有的东西

没有练习、支付、社交、讲题或删号。学习者登录目前只有短信验证码。没有微信或支付宝登录。没有部署配置。CI 只做检查，不发布。知识地图只在导入作业进入 `map_ready` 后写入知识包，没有单独的地图 HTTP 接口。

平台契约正文在 `docs/contracts/`。交付留痕在 `docs/delivery/`。
