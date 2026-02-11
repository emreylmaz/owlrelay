import type { RelayMessage, ExtensionMessage, ConnectionState } from '../shared/types';
import { HEARTBEAT_INTERVAL, RECONNECT_DELAY_BASE, RECONNECT_DELAY_MAX, MAX_RECONNECT_ATTEMPTS } from '../shared/constants';
import { handleRelayMessage } from './commands';
import { getAttachedTabsForRelay } from './tabs';

let socket: WebSocket | null = null;
let heartbeatInterval: ReturnType<typeof setInterval> | null = null;
let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;
let reconnectAttempts = 0;
let currentRelayUrl = '';
let currentToken = '';

// Connection state
let connectionState: ConnectionState = {
  status: 'disconnected',
};

// State change callbacks
const stateListeners: Set<(state: ConnectionState) => void> = new Set();

export function onConnectionStateChange(callback: (state: ConnectionState) => void): () => void {
  stateListeners.add(callback);
  return () => stateListeners.delete(callback);
}

function notifyStateChange(): void {
  for (const listener of stateListeners) {
    listener(connectionState);
  }
  updateBadge();
}

function updateBadge(): void {
  const color = connectionState.status === 'connected' ? '#22c55e' : '#ef4444';
  const text = connectionState.status === 'connected' ? 'ON' : '';
  
  chrome.action.setBadgeBackgroundColor({ color });
  chrome.action.setBadgeText({ text });
}

export function getConnectionState(): ConnectionState {
  return { ...connectionState };
}

export function connect(relayUrl: string, token: string): void {
  // Disconnect any existing connection
  disconnect();
  
  currentRelayUrl = relayUrl;
  currentToken = token;
  reconnectAttempts = 0;
  
  doConnect();
}

function doConnect(): void {
  if (!currentRelayUrl || !currentToken) {
    connectionState = { status: 'error', error: 'Missing relay URL or token' };
    notifyStateChange();
    return;
  }
  
  connectionState = { status: 'connecting' };
  notifyStateChange();
  
  try {
    // Build WebSocket URL with token
    const wsUrl = new URL(currentRelayUrl);
    // Ensure /ws path
    if (!wsUrl.pathname.endsWith('/ws')) {
      wsUrl.pathname = wsUrl.pathname.replace(/\/?$/, '/ws');
    }
    wsUrl.searchParams.set('token', currentToken);
    
    socket = new WebSocket(wsUrl.toString());
    
    socket.onopen = handleOpen;
    socket.onclose = handleClose;
    socket.onerror = handleError;
    socket.onmessage = handleMessage;
  } catch (err) {
    connectionState = { 
      status: 'error', 
      error: err instanceof Error ? err.message : 'Failed to connect' 
    };
    notifyStateChange();
    scheduleReconnect();
  }
}

function handleOpen(): void {
  console.log('[OwlRelay] WebSocket connected');
  reconnectAttempts = 0;
  
  // Start heartbeat
  startHeartbeat();
  
  // Send initial tab list
  sendAttachedTabs();
}

function handleClose(event: CloseEvent): void {
  console.log('[OwlRelay] WebSocket closed:', event.code, event.reason);
  
  stopHeartbeat();
  socket = null;
  
  if (connectionState.status !== 'disconnected') {
    connectionState = { 
      status: 'error', 
      error: `Connection closed: ${event.reason || 'Unknown reason'}` 
    };
    notifyStateChange();
    scheduleReconnect();
  }
}

function handleError(event: Event): void {
  console.error('[OwlRelay] WebSocket error:', event);
}

function handleMessage(event: MessageEvent): void {
  try {
    const message = JSON.parse(event.data) as RelayMessage;
    
    switch (message.type) {
      case 'connect_ack':
        connectionState = {
          status: 'connected',
          sessionId: message.sessionId,
          lastHeartbeat: Date.now(),
        };
        notifyStateChange();
        break;
        
      case 'connect_error':
        connectionState = {
          status: 'error',
          error: message.message,
        };
        notifyStateChange();
        // Don't reconnect on auth errors
        if (message.code === 'INVALID_TOKEN' || message.code === 'TOKEN_EXPIRED') {
          disconnect();
        }
        break;
        
      case 'ping':
        sendMessage({
          type: 'pong',
          timestamp: message.timestamp,
          tabCount: getAttachedTabsForRelay().length,
        });
        connectionState.lastHeartbeat = Date.now();
        break;
        
      case 'command':
        handleRelayMessage(message);
        break;
    }
  } catch (err) {
    console.error('[OwlRelay] Failed to parse message:', err);
  }
}

export function sendMessage(message: ExtensionMessage): void {
  if (socket?.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify(message));
  }
}

export function disconnect(): void {
  currentToken = '';
  reconnectAttempts = MAX_RECONNECT_ATTEMPTS; // Prevent auto-reconnect
  
  stopHeartbeat();
  clearReconnectTimeout();
  
  if (socket) {
    socket.onclose = null; // Prevent reconnect on intentional close
    socket.close();
    socket = null;
  }
  
  connectionState = { status: 'disconnected' };
  notifyStateChange();
}

function startHeartbeat(): void {
  stopHeartbeat();
  heartbeatInterval = setInterval(() => {
    if (socket?.readyState === WebSocket.OPEN) {
      sendMessage({
        type: 'pong',
        timestamp: Date.now(),
        tabCount: getAttachedTabsForRelay().length,
      });
    }
  }, HEARTBEAT_INTERVAL);
}

function stopHeartbeat(): void {
  if (heartbeatInterval) {
    clearInterval(heartbeatInterval);
    heartbeatInterval = null;
  }
}

function scheduleReconnect(): void {
  if (reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
    console.log('[OwlRelay] Max reconnect attempts reached');
    return;
  }
  
  clearReconnectTimeout();
  
  const delay = Math.min(
    RECONNECT_DELAY_BASE * Math.pow(2, reconnectAttempts),
    RECONNECT_DELAY_MAX
  );
  
  console.log(`[OwlRelay] Reconnecting in ${delay}ms (attempt ${reconnectAttempts + 1})`);
  
  reconnectTimeout = setTimeout(() => {
    reconnectAttempts++;
    doConnect();
  }, delay);
}

function clearReconnectTimeout(): void {
  if (reconnectTimeout) {
    clearTimeout(reconnectTimeout);
    reconnectTimeout = null;
  }
}

async function sendAttachedTabs(): Promise<void> {
  const tabs = getAttachedTabsForRelay();
  for (const tab of tabs) {
    sendMessage({
      type: 'tab_attach',
      tabId: tab.uuid,
      url: tab.url,
      title: tab.title,
      favIconUrl: tab.favIconUrl,
    });
  }
}

export function isConnected(): boolean {
  return connectionState.status === 'connected';
}
