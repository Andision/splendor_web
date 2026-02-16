import type { ApiError, GameAction, Room } from "./types";

function resolveApiBase(): string {
  const configured = import.meta.env.VITE_API_BASE;
  if (configured && configured.trim() !== "") {
    return configured;
  }

  if (typeof window !== "undefined") {
    const { protocol, hostname } = window.location;
    return `${protocol}//${hostname}:8080`;
  }

  return "http://localhost:8080";
}

const API_BASE = resolveApiBase();

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {})
    },
    ...init
  });

  if (!response.ok) {
    let err: ApiError | null = null;
    try {
      err = (await response.json()) as ApiError;
    } catch {
      err = null;
    }
    throw new Error(err?.message ?? `HTTP ${response.status}`);
  }

  return (await response.json()) as T;
}

export type CreateRoomResult = {
  room: Room;
  player: {
    id: string;
    name: string;
  };
};

export type JoinRoomResult = {
  room: Room;
  player: {
    id: string;
    name: string;
  };
};

export function createRoom(hostName: string, turnSeconds: number): Promise<CreateRoomResult> {
  return request<CreateRoomResult>("/api/rooms", {
    method: "POST",
    body: JSON.stringify({ hostName, turnSeconds })
  });
}

export function joinRoom(roomId: string, playerName: string): Promise<JoinRoomResult> {
  return request<JoinRoomResult>(`/api/rooms/${roomId}/join`, {
    method: "POST",
    body: JSON.stringify({ playerName })
  });
}

export function startGame(roomId: string, playerId: string): Promise<Room> {
  return request<Room>(`/api/rooms/${roomId}/start`, {
    method: "POST",
    body: JSON.stringify({ playerId })
  });
}

export function loadRoom(roomId: string): Promise<Room> {
  return request<Room>(`/api/rooms/${roomId}`);
}

export function applyAction(roomId: string, playerId: string, action: GameAction): Promise<Room> {
  return request<Room>(`/api/rooms/${roomId}/actions`, {
    method: "POST",
    body: JSON.stringify({ playerId, action })
  });
}

export function buildWsUrl(roomId: string, playerId: string): string {
  const endpoint = new URL(API_BASE);
  endpoint.protocol = endpoint.protocol === "https:" ? "wss:" : "ws:";
  endpoint.pathname = "/ws";
  endpoint.searchParams.set("roomId", roomId);
  endpoint.searchParams.set("playerId", playerId);
  return endpoint.toString();
}
