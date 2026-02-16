# Splendor Web

自部署多人在线 Splendor（前后端分离，Docker Compose 一键启动）。

## 技术栈

- Frontend: React + Vite + TypeScript + Framer Motion
- Backend: Go + net/http + Gorilla WebSocket

## 快速开始（Docker）

```bash
docker compose up --build
```

启动后：

- Frontend: `http://localhost`
- Backend API: `http://localhost:8080`

## 当前进度

- 后端 MVP（房间、开局、动作校验、状态广播）已完成
- 后端测试已覆盖 `game/lobby/app` 主链路并通过
- 前端已完成联调版页面（开房、加房、开局、提交基础动作、WS 同步）

## 目录

- `frontend/`：前端应用
- `backend/`：Go 后端
- `docs/`：规则或数据资料
