// ===== WebSocket Message Types =====

export interface ConnectAck {
  type: 'connect_ack';
  sessionId: string;
  serverTime: number;
  serverVersion: string;
}

export interface ConnectError {
  type: 'connect_error';
  code: 'INVALID_TOKEN' | 'TOKEN_EXPIRED' | 'RATE_LIMITED' | 'SERVER_ERROR';
  message: string;
}

export interface TabAttach {
  type: 'tab_attach';
  tabId: string;
  url: string;
  title: string;
  favIconUrl?: string;
}

export interface TabDetach {
  type: 'tab_detach';
  tabId: string;
}

export interface TabUpdate {
  type: 'tab_update';
  tabId: string;
  url?: string;
  title?: string;
}

export interface Ping {
  type: 'ping';
  timestamp: number;
}

export interface Pong {
  type: 'pong';
  timestamp: number;
  tabCount: number;
}

// ===== Command Types =====

export interface ClickAction {
  kind: 'click';
  selector?: string;
  coordinates?: { x: number; y: number };
  button?: 'left' | 'right' | 'middle';
  modifiers?: ('ctrl' | 'shift' | 'alt' | 'meta')[];
}

export interface TypeAction {
  kind: 'type';
  selector: string;
  text: string;
  clear?: boolean;
  delay?: number;
}

export interface ScrollAction {
  kind: 'scroll';
  selector?: string;
  direction: 'up' | 'down' | 'left' | 'right';
  amount: number;
}

export interface ScreenshotAction {
  kind: 'screenshot';
  fullPage?: boolean;
  clip?: { x: number; y: number; width: number; height: number };
  quality?: number;
}

export interface SnapshotAction {
  kind: 'snapshot';
  maxDepth?: number;
  maxLength?: number;
  includeStyles?: boolean;
}

export interface NavigateAction {
  kind: 'navigate';
  url: string;
  waitUntil?: 'load' | 'domcontentloaded' | 'networkidle';
}

export type CommandAction =
  | ClickAction
  | TypeAction
  | ScrollAction
  | ScreenshotAction
  | SnapshotAction
  | NavigateAction;

export interface CommandRequest {
  type: 'command';
  id: string;
  action: CommandAction;
  tabId: string;
  timeout: number;
}

export interface CommandResponse {
  type: 'command_response';
  id: string;
  success: boolean;
  result?: unknown;
  error?: {
    code: string;
    message: string;
  };
  timing: {
    received: number;
    completed: number;
  };
}

// ===== Relay Messages =====

export type RelayMessage =
  | ConnectAck
  | ConnectError
  | Ping
  | CommandRequest;

export type ExtensionMessage =
  | TabAttach
  | TabDetach
  | TabUpdate
  | Pong
  | CommandResponse;

// ===== Internal Chrome Message Types =====

export interface AttachedTab {
  tabId: number;
  uuid: string;
  url: string;
  title: string;
  favIconUrl?: string;
  attachedAt: number;
}

export interface ConnectionState {
  status: 'disconnected' | 'connecting' | 'connected' | 'error';
  sessionId?: string;
  error?: string;
  lastHeartbeat?: number;
}

export interface StorageData {
  relayUrl: string;
  token: string;
  attachedTabs: AttachedTab[];
}
