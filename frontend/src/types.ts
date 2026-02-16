export type ApiError = {
  code: string;
  message: string;
};

export type TokenSet = {
  white: number;
  blue: number;
  green: number;
  red: number;
  black: number;
  gold: number;
};

export type Card = {
  id: string;
  tier: number;
  bonus: "white" | "blue" | "green" | "red" | "black";
  points: number;
  cost: TokenSet;
};

export type Noble = {
  id: string;
  points: number;
  requirement: TokenSet;
};

export type Player = {
  id: string;
  name: string;
};

export type PlayerState = {
  id: string;
  name: string;
  tokens: TokenSet;
  bonuses: TokenSet;
  reserved: Card[];
  purchasedCount: number;
  points: number;
  nobles: Noble[];
  isConnected: boolean;
  lastAction: string;
};

export type GameState = {
  status: "playing" | "finished";
  turn: number;
  currentPlayerId: string;
  bank: TokenSet;
  tier1: Card[];
  tier2: Card[];
  tier3: Card[];
  deck1Count: number;
  deck2Count: number;
  deck3Count: number;
  nobles: Noble[];
  players: PlayerState[];
  winnerIds: string[];
  finalRound: boolean;
  finalTurnsLeft: number;
};

export type Room = {
  id: string;
  code: string;
  hostId: string;
  status: "waiting" | "playing" | "finished";
  players: Player[];
  createdAt: string;
  startedAt?: string;
  finishedAt?: string;
  game?: GameState;
};

export type ActionPayload = {
  colors?: string[];
  cardId?: string;
  source?: "tableau" | "reserved";
};

export type GameAction = {
  type: "take_tokens" | "reserve_card" | "buy_card" | "pass";
  payload?: ActionPayload;
};

export type WsSnapshotMessage = {
  type: "room_snapshot";
  reason: string;
  room: Room;
};

export type WsActionErrorMessage = {
  type: "action_error";
  error: string;
};
