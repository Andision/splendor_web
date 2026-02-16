import { useEffect, useMemo, useRef, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { applyAction, buildWsUrl, createRoom, joinRoom, loadRoom, startGame } from "./api";
import type { Card, GameAction, PlayerState, Room, WsActionErrorMessage, WsSnapshotMessage } from "./types";
import "./App.css";

type Session = {
  roomId: string;
  playerId: string;
  playerName: string;
};

const TOKEN_COLORS = ["white", "blue", "green", "red", "black"] as const;

function shortId(id: string): string {
  if (!id) return "-";
  if (id.length <= 6) return id;
  return `${id.slice(0, 4)}...${id.slice(-2)}`;
}

function cardCostText(card: Card): string {
  const parts: string[] = [];
  for (const color of TOKEN_COLORS) {
    const n = card.cost[color];
    if (n > 0) {
      parts.push(`${color[0].toUpperCase()}${n}`);
    }
  }
  return parts.length > 0 ? parts.join(" ") : "Free";
}

export default function App() {
  const [hostName, setHostName] = useState("Alice");
  const [joinName, setJoinName] = useState("Bob");
  const [joinRoomId, setJoinRoomId] = useState("");
  const [session, setSession] = useState<Session | null>(null);
  const [room, setRoom] = useState<Room | null>(null);
  const [statusText, setStatusText] = useState("Ready");
  const [eventLog, setEventLog] = useState<string[]>(["System ready"]);
  const [takeSelection, setTakeSelection] = useState<string[]>([]);
  const wsRef = useRef<WebSocket | null>(null);
  const roomRef = useRef<Room | null>(null);

  const isHost = useMemo(() => {
    if (!room || !session) return false;
    return room.hostId === session.playerId;
  }, [room, session]);
  const canStart = useMemo(() => {
    if (!room) return false;
    return room.status === "waiting" && room.players.length >= 2;
  }, [room]);

  const myPlayerState = useMemo(() => {
    if (!room?.game || !session) return null;
    return (room.game.players ?? []).find((p) => p.id === session.playerId) ?? null;
  }, [room, session]);

  const currentPlayerName = useMemo(() => {
    if (!room?.game) return "-";
    const current = (room.game.players ?? []).find((p) => p.id === room.game?.currentPlayerId);
    return current ? current.name : shortId(room.game.currentPlayerId ?? "");
  }, [room]);

  function appendLog(message: string) {
    setEventLog((prev) => {
      const next = [`${new Date().toLocaleTimeString()} ${message}`, ...prev];
      return next.slice(0, 20);
    });
  }

  useEffect(() => {
    roomRef.current = room;
  }, [room]);

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
          setRoom(message.room);
          setStatusText(`Realtime update: ${message.reason}`);
          appendLog(`Snapshot ${message.reason}`);
        } else if (raw.type === "action_error") {
          const err = raw as WsActionErrorMessage;
          setStatusText(`Action rejected: ${err.error}`);
          appendLog(`Action rejected: ${err.error}`);
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
      const result = await createRoom(hostName.trim());
      setSession({
        roomId: result.room.code ?? result.room.id,
        playerId: result.player.id,
        playerName: result.player.name
      });
      setRoom(result.room);
      setJoinRoomId(result.room.code ?? result.room.id);
      setStatusText(`Room ${(result.room.code ?? result.room.id)} created`);
      appendLog(`Room ${(result.room.code ?? result.room.id)} created`);
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Create failed: ${String(err)}`);
    }
  }

  async function onJoinRoom() {
    try {
      const result = await joinRoom(joinRoomId.trim(), joinName.trim());
      setSession({
        roomId: result.room.code ?? result.room.id,
        playerId: result.player.id,
        playerName: result.player.name
      });
      setRoom(result.room);
      setStatusText(`Joined room ${(result.room.code ?? result.room.id)}`);
      appendLog(`Joined room ${(result.room.code ?? result.room.id)}`);
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Join failed: ${String(err)}`);
    }
  }

  async function onStartGame() {
    if (!session || !room) return;
    try {
      const updated = await startGame(room.id, session.playerId);
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

  async function submitAction(action: GameAction) {
    if (!session || !room) return;
    try {
      const updated = await applyAction(room.id, session.playerId, action);
      setRoom(updated);
      setStatusText(`Action sent: ${action.type}`);
      appendLog(`Action ${action.type}`);
    } catch (err) {
      setStatusText(String(err));
      appendLog(`Action failed: ${String(err)}`);
    }
  }

  function toggleTakeColor(color: string) {
    setTakeSelection((prev) => {
      if (prev.includes(color)) {
        return prev.filter((c) => c !== color);
      }
      if (prev.length >= 3) return prev;
      return [...prev, color];
    });
  }

  function clearTakeSelection() {
    setTakeSelection([]);
  }

  function submitTakeSelection() {
    if (takeSelection.length === 0) return;
    void submitAction({ type: "take_tokens", payload: { colors: takeSelection } });
    clearTakeSelection();
  }

  function submitTakeTwoSame(color: string) {
    void submitAction({ type: "take_tokens", payload: { colors: [color, color] } });
  }

  function renderCard(card: Card, owner?: PlayerState) {
    return (
      <motion.div
        key={card.id}
        className="card"
        layout
        initial={{ rotateY: 90, opacity: 0 }}
        animate={{ rotateY: 0, opacity: 1 }}
        whileHover={{ y: -4, scale: 1.02 }}
        transition={{ duration: 0.25 }}
      >
        <strong>{card.id}</strong>
        <span>PV {card.points}</span>
        <span>Bonus {card.bonus}</span>
        <span className="cost">{cardCostText(card)}</span>
        <div className="card-actions">
          {!owner && (
            <>
              <button onClick={() => submitAction({ type: "reserve_card", payload: { cardId: card.id } })}>Reserve</button>
              <button onClick={() => submitAction({ type: "buy_card", payload: { cardId: card.id, source: "tableau" } })}>Buy</button>
            </>
          )}
          {owner && (
            <button onClick={() => submitAction({ type: "buy_card", payload: { cardId: card.id, source: "reserved" } })}>
              Buy Reserved
            </button>
          )}
        </div>
      </motion.div>
    );
  }

  const myReserved = myPlayerState?.reserved ?? [];

  return (
    <div className="page">
      <motion.header
        className="hero"
        initial={{ opacity: 0, y: -12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35 }}
      >
        <h1>Splendor Multiplayer</h1>
        <p>Local self-host build for your private game nights.</p>
      </motion.header>

      {!session ? (
        <section className="panel grid-two">
          <div>
            <h2>Create Room</h2>
            <input value={hostName} onChange={(e) => setHostName(e.target.value)} placeholder="Host name" />
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
          <section className="panel">
            <div className="row">
              <div>
                <h2>Room {session.roomId}</h2>
                <p>You are {session.playerName}</p>
              </div>
              <div className="actions">
                <button onClick={onRefreshRoom}>Refresh</button>
                {isHost && room?.status === "waiting" && (
                  <button onClick={onStartGame} disabled={!canStart} title={canStart ? "Start game" : "Need at least 2 players"}>
                    Start Game
                  </button>
                )}
              </div>
            </div>
            <p className="status">{statusText}</p>
            {room?.status === "waiting" && room.players.length < 2 && (
              <p className="hint">Need at least 2 players to start. Share room ID: {room.code ?? room.id}</p>
            )}
          </section>

          <section className="panel">
            <h3>Players</h3>
            <div className="player-list">
              {room?.players.map((p) => (
                <motion.div
                  key={p.id}
                  className={`player-chip ${room.game?.currentPlayerId === p.id ? "is-current" : ""}`}
                  initial={{ opacity: 0, scale: 0.92 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ duration: 0.22 }}
                >
                  {p.name}
                </motion.div>
              ))}
            </div>
          </section>

          {room?.game && (
            <section className="panel">
              <h3>Game Controls</h3>
              <div className="control-grid">
                <div className="take-builder">
                  <div className="take-pairs">
                    {TOKEN_COLORS.map((color) => (
                      <button key={`pair-${color}`} className="take-pair-btn" onClick={() => submitTakeTwoSame(color)}>
                        Take 2 {color}
                      </button>
                    ))}
                  </div>
                  <div className="take-colors">
                    {TOKEN_COLORS.map((color) => (
                      <button
                        key={color}
                        className={`take-color ${takeSelection.includes(color) ? "is-picked" : ""}`}
                        onClick={() => toggleTakeColor(color)}
                      >
                        {color}
                      </button>
                    ))}
                  </div>
                  <div className="take-actions">
                    <button onClick={submitTakeSelection}>Take Selected</button>
                    <button onClick={clearTakeSelection}>Clear</button>
                  </div>
                  <small>3-different picker: {takeSelection.length > 0 ? takeSelection.join(", ") : "none"}</small>
                </div>
                <button onClick={() => submitAction({ type: "pass" })}>Pass</button>
              </div>

              <div className="board-grid">
                <article>
                  <h4>Turn</h4>
                  <p>{room.game.turn}</p>
                  <small>
                    Current: <strong>{currentPlayerName}</strong>
                  </small>
                  {room.game.finalRound && <p className="final-round">Final round: {room.game.finalTurnsLeft} turns left</p>}
                </article>

                <article>
                  <h4>Bank</h4>
                  <motion.div className="tokens" layout>
                    {TOKEN_COLORS.map((color) => (
                      <motion.span key={color} className={`token ${color}`} layout>
                        {color[0].toUpperCase()} {room.game?.bank[color] ?? 0}
                      </motion.span>
                    ))}
                    <motion.span className="token gold" layout>
                      G {room.game?.bank.gold ?? 0}
                    </motion.span>
                  </motion.div>
                </article>

                <article>
                  <h4>Me</h4>
                  {!myPlayerState ? (
                    <p>Not found in room</p>
                  ) : (
                    <div className="my-state">
                      <p>Points: {myPlayerState.points}</p>
                      <p>
                        Tokens: W{myPlayerState.tokens.white} B{myPlayerState.tokens.blue} G{myPlayerState.tokens.green} R
                        {myPlayerState.tokens.red} K{myPlayerState.tokens.black} Gold{myPlayerState.tokens.gold}
                      </p>
                      <p>
                        Bonuses: W{myPlayerState.bonuses.white} B{myPlayerState.bonuses.blue} G{myPlayerState.bonuses.green} R
                        {myPlayerState.bonuses.red} K{myPlayerState.bonuses.black}
                      </p>
                      <p>Reserved: {myReserved.length}</p>
                    </div>
                  )}
                </article>
              </div>

              <div className="tiers">
                <article>
                  <h4>Tier 3 (Deck {room.game.deck3Count ?? 0})</h4>
                  <div className="cards">{(room.game.tier3 ?? []).map((card) => renderCard(card))}</div>
                </article>

                <article>
                  <h4>Tier 2 (Deck {room.game.deck2Count ?? 0})</h4>
                  <div className="cards">{(room.game.tier2 ?? []).map((card) => renderCard(card))}</div>
                </article>

                <article>
                  <h4>Tier 1 (Deck {room.game.deck1Count ?? 0})</h4>
                  <div className="cards">{(room.game.tier1 ?? []).map((card) => renderCard(card))}</div>
                </article>
              </div>

              <article>
                <h4>Reserved Cards</h4>
                {myPlayerState && myReserved.length > 0 ? (
                  <div className="cards">{myReserved.map((card) => renderCard(card, myPlayerState))}</div>
                ) : (
                  <p>No reserved cards</p>
                )}
              </article>

              <article>
                <h4>Player Board</h4>
                <div className="player-board-grid">
                  {(room.game.players ?? []).map((player) => (
                    <motion.div
                      key={player.id}
                      layout
                      className={`player-board-card ${room.game?.currentPlayerId === player.id ? "is-current-turn" : ""}`}
                    >
                      <div className="player-board-head">
                        <strong>{player.name}</strong>
                        <span>{player.isConnected ? "Online" : "Offline"}</span>
                      </div>
                      <p>Points: {player.points}</p>
                      <p>Cards: {player.purchasedCount}</p>
                      <p>
                        Bonus W/B/G/R/K: {player.bonuses.white}/{player.bonuses.blue}/{player.bonuses.green}/{player.bonuses.red}/
                        {player.bonuses.black}
                      </p>
                    </motion.div>
                  ))}
                </div>
              </article>
            </section>
          )}

          <section className="panel">
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
          </section>
        </>
      )}
    </div>
  );
}
