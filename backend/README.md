# Splendor Backend (Go)

后端已完成可联调的 MVP 版本，采用内存态房间与对局状态机。

## 已实现能力

- 房间：创建、加入、查询
- 对局：房主开局、回合推进、终局判定
- 动作校验：
  - `take_tokens`（拿 1-3 个不同色，或 2 个同色）
  - `discard_tokens`（主动弃置任意数量代币回银行）
  - `reserve_card`（预留明牌，最多 3 张，尝试拿 1 金）
  - `buy_card`（购买明牌或预留牌，支持金代币补足）
  - `pass`
- 规则：
  - 2/3/4 人宝石初始数量（4/5/7）
  - 每回合代币上限 10
  - 贵族自动判定领取（每回合最多 1 个）
  - 达到 15 分后触发终局轮，按分数与已购买牌数判胜
- 通信：
  - HTTP API（状态查询与动作提交）
  - WebSocket（房间状态快照广播）

卡牌数据为硬编码常量，已使用 `docs/Splendor Cards.csv` 中的 90 张开发卡（Tier1:40 / Tier2:30 / Tier3:20）。

## API

### 健康检查

- `GET /api/health`

### 房间

- `POST /api/rooms`
  - body: `{ "hostName": "Alice", "turnSeconds": 30 }`
  - `turnSeconds` 可选，默认 `30`，允许范围 `5-300`
- `GET /api/rooms/{roomId}`
- `POST /api/rooms/{roomId}/join`
  - body: `{ "playerName": "Bob" }`
- `POST /api/rooms/{roomId}/start`
  - body: `{ "playerId": "HOST_PLAYER_ID" }`

### 对局状态与动作

- `GET /api/rooms/{roomId}/state`
- `POST /api/rooms/{roomId}/actions`
  - body:

```json
{
  "playerId": "PLAYER_ID",
  "action": {
    "type": "take_tokens",
    "payload": {
      "colors": ["white", "blue", "green"]
    }
  }
}
```

`buy_card` 示例：

```json
{
  "playerId": "PLAYER_ID",
  "action": {
    "type": "buy_card",
    "payload": {
      "cardId": "2_blue_03",
      "source": "tableau"
    }
  }
}
```

## WebSocket

- `GET /ws?roomId=ROOM_ID&playerId=PLAYER_ID`

客户端消息：

- `{"type":"action","action":{...}}`
- `{"type":"ping"}`

服务端消息：

- `room_snapshot`：完整房间快照（`reason` 如 `connected` / `player_joined` / `action_applied`）
- `action_error`
- `pong`

## 本地运行

```bash
go run ./cmd/server
```

默认监听 `:8080`，可通过环境变量覆盖：

```bash
APP_ADDR=:9000 go run ./cmd/server
```

## Docker

```bash
docker build -t splendor-backend:dev .
docker run --rm -p 8080:8080 splendor-backend:dev
```

## 当前限制（MVP）

- 对局状态在内存中，服务重启会丢失
- 还未接入数据库、鉴权与断线重连恢复
- 贵族数据仍为代码内置默认集
