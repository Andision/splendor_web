# Splendor Frontend

前端使用 React + Vite + TypeScript。

## 启动

```bash
npm install
npm run dev
```

默认开发地址：`http://localhost:5173`

## 环境变量

复制 `.env.example` 为 `.env`，可配置：

- `VITE_API_BASE`：后端地址。默认会按当前页面域名自动拼成 `http(s)://<hostname>:8080`

## 当前能力

- 创建房间
- 加入房间
- 房主开局
- 发送基础动作（拿代币、预留、Pass）
- WebSocket 房间快照同步

## Docker

推荐在仓库根目录启动：

```bash
docker compose up --build
```

默认访问：

- 前端：`http://localhost`
- 后端：`http://localhost:8080`
