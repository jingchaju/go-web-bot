# Go Telegram Bot Admin

一个模块化 Go + Gin + Redis + PostgreSQL + GORM + telebot.v4 后端，以及 Vue 3 + TypeScript + Vite 7 + Tailwind CSS + Pinia 后台管理前端。

## 启动流程

1. 复制 `.env.example` 为 `.env`，配置 PostgreSQL、Redis、JWT_SECRET、ADMIN_SECRET。
2. 运行后端：`go run .`。
3. 首次启动会在 `admin_config` 表自动创建管理员，并在标准输出打印一次性账号和密码，请立即保存。
4. 进入前端目录，复制 `web/.env.example` 为 `web/.env`，确保 `VITE_API_HMAC_SECRET` 与后端 `JWT_SECRET` 一致。
5. 安装前端依赖并启动：`npm install && npm run dev`。
6. 登录后台，在 “Telegram Bot 管理” 保存 Bot token、webhook 路径、端口和 secret 后点击启动 Bot。

## 目录

- `internal/config`：读取 `.env` 并集中定义配置。
- `internal/logger`：统一 info/warning/error/debug/fatal 日志。
- `internal/ctime`：中国时间和时间戳。
- `internal/httpclient`：全局 HTTP 客户端与 GET/POST 封装。
- `internal/redisclient`：Redis 客户端和队列封装。
- `internal/db`、`internal/models`、`internal/dao`：PostgreSQL/GORM、表结构和 DAO。
- `internal/bot`：全局 Telegram Bot 管理器、webhook secret 校验、Redis 更新队列和分发。
- `internal/router`、`internal/handlers`、`internal/middleware`：Gin 路由、API 处理器、JWT 和签名校验。
- `web/src`：Vue 后台管理系统，按 api、layout、store、style、view 模块化组织。
