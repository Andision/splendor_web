import { useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { applyAction, buildWsUrl, createRoom, joinRoom, loadRoom, startGame } from "./api";
import type { Card, GameAction, Noble, PlayerState, Room, WsActionErrorMessage, WsSnapshotMessage } from "./types";
import "./App.css";

type Session = {
  roomId: string;
  playerId: string;
  playerName: string;
};

type ToastMessage = {
  id: number;
  text: string;
};

const TOKEN_COLORS = ["white", "blue", "green", "red", "black"] as const;
const TOKEN_ACTION_COLORS = ["black", "blue", "white", "green", "red", "gold"] as const;
const PLAYER_STAT_COLORS = ["black", "blue", "white", "green", "red"] as const;
const SESSION_STORAGE_KEY = "splendor_session_v1";
const BONUS_LABEL: Record<string, string> = {
  white: "W",
  blue: "B",
  green: "G",
  red: "R",
  black: "K"
};
const GEM_BG: Record<string, string> = {
  white: "linear-gradient(135deg, #f3f5f9, #d9dee8)",
  blue: "linear-gradient(135deg, #7ebcff, #2f6ec6)",
  green: "linear-gradient(135deg, #66d6a3, #258a5f)",
  red: "linear-gradient(135deg, #ff8b7d, #b8403f)",
  black: "linear-gradient(135deg, #8a93a4, #2f3440)"
};
const CARD_BG: Record<string, string> = {
  white: "linear-gradient(165deg, #dfe5ef 0%, #b7c0cf 55%, #8f9caf 100%)",
  blue: "linear-gradient(165deg, #b6dbff 0%, #6da8e6 55%, #3f6ea6 100%)",
  green: "linear-gradient(165deg, #c9f4db 0%, #6bc59a 55%, #2f6f56 100%)",
  red: "linear-gradient(165deg, #ffd1ca 0%, #e78978 55%, #974a45 100%)",
  black: "linear-gradient(165deg, #c8cfdb 0%, #858ea0 55%, #454b58 100%)"
};

function shortId(id: string): string {
  if (!id) return "-";
  if (id.length <= 6) return id;
  return `${id.slice(0, 4)}...${id.slice(-2)}`;
}

function tokenParts(tokenSet: Record<string, number>): Array<{ color: string; amount: number }> {
  const parts: Array<{ color: string; amount: number }> = [];
  for (const color of TOKEN_COLORS) {
    const n = tokenSet[color];
    if (n > 0) {
      parts.push({ color, amount: n });
    }
  }
  return parts;
}

export default function App() {
  const STAGE_WIDTH = 1600;
  const STAGE_HEIGHT = 1000;
  const STAGE_PADDING = 16;
  const [hostName, setHostName] = useState("Alice");
  const [createTurnSeconds, setCreateTurnSeconds] = useState(30);
  const [joinName, setJoinName] = useState("Bob");
  const [joinRoomId, setJoinRoomId] = useState("");
  const [session, setSession] = useState<Session | null>(null);
  const [room, setRoom] = useState<Room | null>(null);
  const [statusText, setStatusText] = useState("Ready");
  const [eventLog, setEventLog] = useState<string[]>(["System ready"]);
  const [toasts, setToasts] = useState<ToastMessage[]>([]);
  const [tokenDraft, setTokenDraft] = useState<Record<string, number>>({
    white: 0,
    blue: 0,
    green: 0,
    red: 0,
    black: 0,
    gold: 0
  });
  const [stageScale, setStageScale] = useState(1);
  const [turnCountdown, setTurnCountdown] = useState(0);
  const wsRef = useRef<WebSocket | null>(null);
  const roomRef = useRef<Room | null>(null);
  const toastIdRef = useRef(1);

  const isHost = useMemo(() => {
    if (!room || !session) return false;
    return room.hostId === session.playerId;
  }, [room, session]);

  const canStart = useMemo(() => {
    if (!room) return false;
    return room.status === "waiting" && room.players.length >= 2;
  }, [room]);

  const currentPlayerName = useMemo(() => {
    if (!room?.game) return "-";
    const current = (room.game.players ?? []).find((p) => p.id === room.game?.currentPlayerId);
    return current ? current.name : shortId(room.game.currentPlayerId ?? "");
  }, [room]);

  const myPlayerState = useMemo(() => {
    if (!room?.game || !session) return null;
    return (room.game.players ?? []).find((p) => p.id === session.playerId) ?? null;
  }, [room, session]);

  function appendLog(message: string) {
    setEventLog((prev) => {
      const next = [`${new Date().toLocaleTimeString([], { hour12: false })} ${message}`, ...prev];
      return next.slice(0, 120);
    });
  }

  function dismissToast(id: number) {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }

  function pushErrorToast(text: string) {
    const id = toastIdRef.current++;
    setToasts((prev) => [...prev, { id, text }]);
    window.setTimeout(() => {
      dismissToast(id);
    }, 4500);
  }

  function goHome() {
    setSession(null);
    setRoom(null);
    setJoinRoomId("");
    setTokenDraft({ white: 0, blue: 0, green: 0, red: 0, black: 0, gold: 0 });
    setStatusText("Ready");
    appendLog("Back to home");
  }

  function playerTokensText(player: PlayerState): string {
    return `W${player.tokens.white} B${player.tokens.blue} G${player.tokens.green} R${player.tokens.red} K${player.tokens.black} Gd${player.tokens.gold}`;
  }

  function totalTokens(player: PlayerState): number {
    return player.tokens.white + player.tokens.blue + player.tokens.green + player.tokens.red + player.tokens.black + player.tokens.gold;
  }

  function totalCards(player: PlayerState): number {
    return player.purchasedCount + (player.nobles?.length ?? 0);
  }

  function applySnapshotLog(nextRoom: Room, reason: string) {
    const prevRoom = roomRef.current;

    if (reason === "player_joined") {
      appendLog(`Player joined. Total players: ${nextRoom.players.length}`);
    }
    if (reason === "game_started") {
      appendLog("Game started");
    }
    if (reason === "player_connected") {
      appendLog("A player connected");
    }
    if (reason === "player_disconnected") {
      appendLog("A player disconnected");
    }

    if (!nextRoom.game) return;

    const prevByID = new Map<string, PlayerState>();
    for (const p of prevRoom?.game?.players ?? []) {
      prevByID.set(p.id, p);
    }

    for (const p of nextRoom.game.players ?? []) {
      const prev = prevByID.get(p.id);
      if (p.lastAction && p.lastAction !== prev?.lastAction) {
        appendLog(`${p.name} used ${p.lastAction} | P:${p.points} | ${playerTokensText(p)}`);
      }
    }

    if ((prevRoom?.game?.turn ?? 0) !== nextRoom.game.turn) {
      const current = (nextRoom.game.players ?? []).find((p) => p.id === nextRoom.game?.currentPlayerId);
      appendLog(`Turn ${nextRoom.game.turn} -> ${current?.name ?? shortId(nextRoom.game.currentPlayerId ?? "")}`);
    }
  }

  useEffect(() => {
    const updateStageScale = () => {
      const availWidth = Math.max(window.innerWidth - STAGE_PADDING, 320);
      const availHeight = Math.max(window.innerHeight - STAGE_PADDING, 180);
      const nextScale = Math.min(availWidth / STAGE_WIDTH, availHeight / STAGE_HEIGHT);
      setStageScale(Number.isFinite(nextScale) && nextScale > 0 ? nextScale : 1);
    };
    updateStageScale();
    window.addEventListener("resize", updateStageScale);
    return () => window.removeEventListener("resize", updateStageScale);
  }, []);

  useEffect(() => {
    roomRef.current = room;
  }, [room]);

  useEffect(() => {
    try {
      const raw = localStorage.getItem(SESSION_STORAGE_KEY);
      if (!raw) return;
      const restored = JSON.parse(raw) as Session;
      if (!restored?.roomId || !restored?.playerId) return;
      setSession(restored);
      setJoinRoomId(restored.roomId);
      appendLog(`Restored session for room ${restored.roomId}`);
    } catch {
      // ignore malformed storage
    }
  }, []);

  useEffect(() => {
    if (!session) {
      localStorage.removeItem(SESSION_STORAGE_KEY);
      return;
    }
    localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(session));
  }, [session]);

  useEffect(() => {
    if (!session) return;
    void (async () => {
      try {
        const latest = await loadRoom(session.roomId);
        setRoom(latest);
      } catch {
        setSession(null);
        setRoom(null);
        setStatusText("Session expired, please rejoin room");
        appendLog("Stored session is no longer valid");
      }
    })();
  }, [session]);

  useEffect(() => {
    if (!room?.game || !room.turnDeadline) {
      setTurnCountdown(0);
      return;
    }

    const deadline = new Date(room.turnDeadline).getTime();
    const tick = () => {
      const remain = Math.max(0, Math.ceil((deadline - Date.now()) / 1000));
      setTurnCountdown(remain);
    };
    tick();
    const timer = window.setInterval(tick, 200);
    return () => window.clearInterval(timer);
  }, [room?.game, room?.turnDeadline]);

  useEffect(() => {
    if (!session) return;

    const ws = new WebSocket(buildWsUrl(session.roomId, session.playerId));
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const raw = JSON.parse(event.data) as WsSnapshotMessage | WsActionErrorMessage | { type: string };
        if (raw.type === "room_snapshot") {
          const message = raw as WsSnapshotMessage;
          if (roomRef.current?.game && !message.room.game) {
            appendLog(`Ignored stale snapshot: ${message.reason}`);
            return;
          }
          applySnapshotLog(message.room, message.reason);
          setRoom(message.room);
          setStatusText(`Realtime update: ${message.reason}`);
        } else if (raw.type === "action_error") {
          const err = raw as WsActionErrorMessage;
          setStatusText(`Action rejected: ${err.error}`);
          appendLog(`Action rejected: ${err.error}`);
          pushErrorToast(`Action rejected: ${err.error}`);
        }
      } catch {
        setStatusText("Received unknown WS payload");
        appendLog("Unknown WS payload");
      }
    };

    ws.onopen = () => {
      setStatusText("WebSocket connected");
      appendLog("WS connected");
    };

    ws.onclose = () => {
      setStatusText("WebSocket disconnected");
      appendLog("WS disconnected");
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [session]);

  async function onCreateRoom() {
    try {
      const clampedSeconds = Math.max(5, Math.min(300, Number(createTurnSeconds) || 30));
      const result = await createRoom(hostName.trim(), clampedSeconds);
      const ref = result.room.code ?? result.room.id;
      setSession({ roomId: ref, playerId: result.player.id, playerName: result.player.name });
      setRoom(result.room);
      setJoinRoomId(ref);
      setStatusText(`Room ${ref} created`);
      appendLog(`Room ${ref} created`);
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Create failed: ${String(err)}`);
    }
  }

  async function onJoinRoom() {
    try {
      const result = await joinRoom(joinRoomId.trim(), joinName.trim());
      const ref = result.room.code ?? result.room.id;
      setSession({ roomId: ref, playerId: result.player.id, playerName: result.player.name });
      setRoom(result.room);
      setStatusText(`Joined room ${ref}`);
      appendLog(`Joined room ${ref}`);
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Join failed: ${String(err)}`);
    }
  }

  async function onStartGame() {
    if (!session || !room) return;
    try {
      const updated = await startGame(session.roomId, session.playerId);
      setRoom(updated);
      setStatusText("Game started");
      appendLog("Game started");
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Start failed: ${String(err)}`);
    }
  }

  async function onRefreshRoom() {
    if (!session) return;
    try {
      const latest = await loadRoom(session.roomId);
      setRoom(latest);
      setStatusText("Room refreshed");
      appendLog("Room refreshed");
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Refresh failed: ${String(err)}`);
    }
  }

  async function submitAction(action: GameAction): Promise<boolean> {
    if (!session) return false;
    try {
      const updated = await applyAction(session.roomId, session.playerId, action);
      setRoom(updated);
      setStatusText(`Action sent: ${action.type}`);
      appendLog(`Action ${action.type}`);
      return true;
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Action failed: ${String(err)}`);
      pushErrorToast(`Action failed: ${String(err)}`);
      return false;
    }
  }

  function adjustTokenDraft(color: string, delta: number) {
    setTokenDraft((prev) => {
      const current = prev[color] ?? 0;
      const bank = room?.game?.bank[color as keyof typeof room.game.bank] ?? 0;
      const owned = myPlayerState?.tokens[color as keyof typeof myPlayerState.tokens] ?? 0;
      const min = Math.max(-2, -owned);
      const max = Math.min(2, bank);
      const next = Math.max(min, Math.min(max, current + delta));
      return { ...prev, [color]: next };
    });
  }

  async function submitTokenDraft() {
    const positives: string[] = [];
    const negatives: string[] = [];
    const adjust: Record<string, number> = {};

    for (const color of TOKEN_ACTION_COLORS) {
      const amount = tokenDraft[color] ?? 0;
      if (amount !== 0) {
        adjust[color] = amount;
      }
      if (amount > 0) {
        for (let i = 0; i < amount; i++) positives.push(color);
      } else if (amount < 0) {
        for (let i = 0; i < Math.abs(amount); i++) negatives.push(color);
      }
    }

    if (positives.length === 0 && negatives.length === 0) {
      setStatusText("No token change selected");
      appendLog("No token change selected");
      return;
    }
    if (positives.includes("gold")) {
      setStatusText("Cannot take gold directly");
      appendLog("Cannot take gold directly");
      return;
    }

    let ok = false;
    if (positives.length > 0 && negatives.length > 0) {
      ok = await submitAction({ type: "adjust_tokens", payload: { adjust } });
    } else if (positives.length > 0) {
      ok = await submitAction({ type: "take_tokens", payload: { colors: positives } });
    } else {
      ok = await submitAction({ type: "discard_tokens", payload: { colors: negatives } });
    }

    if (ok) {
      setTokenDraft({ white: 0, blue: 0, green: 0, red: 0, black: 0, gold: 0 });
    }
  }

  function renderNoble(noble: Noble) {
    const reqs = tokenParts(noble.requirement);
    return (
      <div key={noble.id} className="noble-card">
        <div className="noble-head">
          <strong>Noble</strong>
          <span>{noble.points} PV</span>
        </div>
        <div className="noble-reqs">
          {reqs.map((item) => (
            <span key={`${noble.id}-${item.color}`} className={`cost-chip ${item.color}`}>
              {item.amount}
            </span>
          ))}
        </div>
      </div>
    );
  }

  function renderCard(card: Card, owner?: PlayerState, readonly = false) {
    const costs = tokenParts(card.cost);
    const cardStyle = {
      "--card-bg": CARD_BG[card.bonus],
      "--gem-bg": GEM_BG[card.bonus]
    } as CSSProperties;
    const buyAction = owner
      ? { type: "buy_card" as const, payload: { cardId: card.id, source: "reserved" as const } }
      : { type: "buy_card" as const, payload: { cardId: card.id, source: "tableau" as const } };
    const canBuyReserved = Boolean(owner) && owner?.id === session?.playerId;

    return (
      <motion.div
        key={card.id}
        className="card"
        style={cardStyle}
        layout
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
      >
        <div className="card-top">
          <span className="card-points">{card.points}</span>
          <span className={`card-gem ${card.bonus}`}>{BONUS_LABEL[card.bonus] ?? card.bonus[0].toUpperCase()}</span>
        </div>
        <div className="card-body" />
        <div className="card-bottom">
          <div className="card-costs card-costs-vertical">
            {costs.length === 0 ? <span className="free-chip">Free</span> : null}
            {costs.map((item) => (
              <span key={`${card.id}-${item.color}`} className={`cost-chip ${item.color}`}>
                {item.amount}
              </span>
            ))}
          </div>
          <div className="card-actions card-actions-vertical">
            <button disabled={Boolean(owner) && !canBuyReserved} onClick={() => submitAction(buyAction)}>
              Buy
            </button>
            <button disabled={Boolean(owner) || readonly} onClick={() => submitAction({ type: "reserve_card", payload: { cardId: card.id } })}>
              Reserve
            </button>
          </div>
        </div>
      </motion.div>
    );
  }

  function renderReservedMini(card: Card, ownerId: string) {
    const costs = tokenParts(card.cost);
    const canBuy = ownerId === session?.playerId;
    return (
      <div key={`mini-${ownerId}-${card.id}`} className="reserved-mini">
        <div className="reserved-mini-head">
          <span>{card.points} PV</span>
          <span className={`cost-chip ${card.bonus}`}>{BONUS_LABEL[card.bonus] ?? card.bonus[0].toUpperCase()}</span>
        </div>
        <div className="reserved-mini-costs">
          {costs.slice(0, 5).map((item) => (
            <span key={`mini-cost-${card.id}-${item.color}`} className={`cost-chip ${item.color}`}>
              {item.amount}
            </span>
          ))}
        </div>
        <button disabled={!canBuy} onClick={() => submitAction({ type: "buy_card", payload: { cardId: card.id, source: "reserved" } })}>
          Buy
        </button>
      </div>
    );
  }

  return (
    <div className="viewport">
      <div className="page" style={{ "--stage-scale": String(stageScale) } as CSSProperties}>
        <div className="toast-stack">
          {toasts.map((toast) => (
            <button key={toast.id} className="toast-error" onClick={() => dismissToast(toast.id)} title="Click to dismiss">
              {toast.text}
            </button>
          ))}
        </div>
        <motion.header className="hero" initial={{ opacity: 0, y: -8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.3 }}>
          <h1>Splendor Multiplayer</h1>
        </motion.header>

        {!session ? (
          <section className="panel grid-two">
            <div>
              <h2>Create Room</h2>
              <input value={hostName} onChange={(e) => setHostName(e.target.value)} placeholder="Host name" />
              <input
                type="number"
                min={5}
                max={300}
                value={createTurnSeconds}
                onChange={(e) => setCreateTurnSeconds(Number(e.target.value))}
                placeholder="Turn seconds"
              />
              <button onClick={onCreateRoom}>Create</button>
            </div>
            <div>
              <h2>Join Room</h2>
              <input value={joinRoomId} onChange={(e) => setJoinRoomId(e.target.value)} placeholder="Room ID" />
              <input value={joinName} onChange={(e) => setJoinName(e.target.value)} placeholder="Player name" />
              <button onClick={onJoinRoom}>Join</button>
            </div>
          </section>
        ) : (
          <>
            {!room ? (
              <section className="panel waiting-panel">
                <h3>Loading Room</h3>
                <p className="status">Restoring room state...</p>
              </section>
            ) : room.game ? (
              <section className="game-frame">
                <div className="left-sidebar">
                  <article className="panel room-panel">
                    <h3>Room Info</h3>
                    <p>Room: {room.code ?? room.id}</p>
                    <p>Status: {room.status}</p>
                    <p>You: {session.playerName}</p>
                    <p>Your ID: {shortId(session.playerId)}</p>
                    <p>Players: {room.players.length}</p>
                    <p>Turn Seconds: {room.turnSeconds}</p>
                    <p>Live: {statusText}</p>
                    <div className="room-actions">
                      <button onClick={onRefreshRoom}>Refresh</button>
                      {isHost && room?.status === "waiting" && (
                        <button onClick={onStartGame} disabled={!canStart} title={canStart ? "Start game" : "Need at least 2 players"}>
                          Start
                        </button>
                      )}
                      <button onClick={goHome}>Back Home</button>
                    </div>
                    {room?.status === "waiting" && room.players.length < 2 && <p className="hint">Need at least 2 players to start.</p>}
                  </article>

                  <article className="panel log-panel">
                    <h3>Event Log</h3>
                    <div className="event-log">
                      <AnimatePresence>
                        {eventLog.map((line, index) => (
                          <motion.p key={`${line}-${index}`} initial={{ opacity: 0, x: 8 }} animate={{ opacity: 1, x: 0 }} exit={{ opacity: 0 }}>
                            {line}
                          </motion.p>
                        ))}
                      </AnimatePresence>
                    </div>
                  </article>
                </div>

                <div className="cards-column">
                  <article className="panel cards-panel">
                    <h3>Cards</h3>
                    <div className="nobles-row">
                      <h4>Nobles</h4>
                      <div className="nobles">{(room.game.nobles ?? []).map((noble) => renderNoble(noble))}</div>
                    </div>
                    <div className="tiers">
                      <article>
                        <h4>Tier 1 ({room.game.deck1Count ?? 0})</h4>
                        <div className="cards">{(room.game.tier1 ?? []).map((card) => renderCard(card))}</div>
                      </article>
                      <article>
                        <h4>Tier 2 ({room.game.deck2Count ?? 0})</h4>
                        <div className="cards">{(room.game.tier2 ?? []).map((card) => renderCard(card))}</div>
                      </article>
                      <article>
                        <h4>Tier 3 ({room.game.deck3Count ?? 0})</h4>
                        <div className="cards">{(room.game.tier3 ?? []).map((card) => renderCard(card))}</div>
                      </article>
                    </div>
                  </article>
                </div>

                <div className="right-column">
                  <article className="panel tokens-panel">
                    <div className="turn-inline">
                      <span>Turn {room.game.turn}</span>
                      <span>Current: {currentPlayerName}</span>
                      <span>Timer: {turnCountdown}s / {room.turnSeconds}s</span>
                      {room.game.finalRound && <span className="final-round-inline">Final round: {room.game.finalTurnsLeft}</span>}
                    </div>

                    <div className="token-picker">
                      {TOKEN_ACTION_COLORS.map((color) => (
                        <div key={`picker-${color}`} className="token-picker-item">
                          <div className={`token-coin ${color}`}>
                            <span className="token-coin-badge">{room.game?.bank[color as keyof typeof room.game.bank] ?? 0}</span>
                          </div>
                          {color !== "gold" && (
                            <div className="token-adjust">
                              <button onClick={() => adjustTokenDraft(color, -1)}>−</button>
                              <span>{tokenDraft[color] ?? 0}</span>
                              <button onClick={() => adjustTokenDraft(color, 1)}>+</button>
                            </div>
                          )}
                          {color === "gold" && <div className="token-adjust-placeholder" />}
                        </div>
                      ))}
                      <div className="token-picker-actions">
                        <button onClick={() => void submitTokenDraft()}>Submit Selection</button>
                        <button
                          onClick={() => setTokenDraft({ white: 0, blue: 0, green: 0, red: 0, black: 0, gold: 0 })}
                        >
                          Clear
                        </button>
                        <button onClick={() => void submitAction({ type: "pass" })}>Pass</button>
                      </div>
                    </div>
                  </article>

                  <article className="panel players-panel">
                    <div className="player-board-grid">
                      {(room.game.players ?? []).map((player) => (
                        <div key={player.id} className={`player-board-card ${room.game?.currentPlayerId === player.id ? "is-current-turn" : ""}`}>
                          <div className="player-board-content">
                            <div className="player-info-main">
                              <div className="player-board-head">
                                <strong>{player.name}</strong>
                                <span>{player.isConnected ? "Online" : "Offline"}</span>
                              </div>
                              <div className="player-summary-stats">
                                <span>Score: {player.points}</span>
                                <span>Tokens: {totalTokens(player)}</span>
                                <span>Cards: {totalCards(player)}</span>
                              </div>
                              <div className="player-gem-stats">
                                {PLAYER_STAT_COLORS.map((color) => (
                                  <div key={`${player.id}-${color}`} className={`player-gem-box ${color}`}>
                                    <strong>
                                      {player.tokens[color as keyof typeof player.tokens]}/{player.bonuses[color as keyof typeof player.bonuses]}
                                    </strong>
                                    <small>T/B</small>
                                  </div>
                                ))}
                                <div className="player-gem-box gold">
                                  <strong>{player.tokens.gold}</strong>
                                  <small>Gold</small>
                                </div>
                              </div>
                            </div>
                            <aside className="player-reserved-side">
                              <p className="reserved-side-title">Reserved {(player.reserved ?? []).length}/3</p>
                              {(player.reserved ?? []).length > 0 ? (
                                <div className="reserved-mini-list">{(player.reserved ?? []).map((card) => renderReservedMini(card, player.id))}</div>
                              ) : (
                                <p className="reserved-empty-inline">None</p>
                              )}
                            </aside>
                          </div>
                        </div>
                      ))}
                    </div>
                  </article>
                </div>
              </section>
            ) : (
              <section className="panel waiting-panel">
                <h3>Waiting Room</h3>
                <p>Room ID: {room.code ?? room.id}</p>
                <p>Turn Seconds: {room.turnSeconds}</p>
                <div className="room-actions">
                  <button onClick={onRefreshRoom}>Refresh</button>
                  {isHost && room?.status === "waiting" && (
                    <button onClick={onStartGame} disabled={!canStart} title={canStart ? "Start game" : "Need at least 2 players"}>
                      Start Game
                    </button>
                  )}
                  <button onClick={goHome}>Back Home</button>
                </div>
                <p className="status">{statusText}</p>
                <div className="player-list">
                  {room?.players.map((p) => (
                    <span className="player-chip" key={p.id}>
                      {p.name}
                    </span>
                  ))}
                </div>
              </section>
            )}
          </>
        )}
      </div>
    </div>
  );
}
