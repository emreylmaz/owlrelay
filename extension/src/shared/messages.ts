import type { ConnectionState, AttachedTab, CommandAction } from './types';

// ===== Background ↔ Popup Messages =====

export type PopupToBackgroundMessage =
  | { type: 'GET_STATUS' }
  | { type: 'CONNECT'; relayUrl: string; token: string }
  | { type: 'DISCONNECT' }
  | { type: 'ATTACH_TAB'; tabId: number }
  | { type: 'DETACH_TAB'; tabId: number }
  | { type: 'GET_ATTACHED_TABS' };

export type BackgroundToPopupResponse =
  | { type: 'STATUS'; state: ConnectionState; attachedTabs: AttachedTab[] }
  | { type: 'CONNECTED'; sessionId: string }
  | { type: 'DISCONNECTED' }
  | { type: 'ERROR'; message: string }
  | { type: 'TAB_ATTACHED'; tab: AttachedTab }
  | { type: 'TAB_DETACHED'; tabId: number }
  | { type: 'ATTACHED_TABS'; tabs: AttachedTab[] };

// ===== Background ↔ Content Script Messages =====

export type BackgroundToContentMessage =
  | { type: 'EXECUTE_COMMAND'; commandId: string; action: CommandAction }
  | { type: 'TAKE_SCREENSHOT'; commandId: string }
  | { type: 'GET_SNAPSHOT'; commandId: string; maxDepth?: number; maxLength?: number };

export type ContentToBackgroundMessage =
  | { type: 'COMMAND_RESULT'; commandId: string; success: boolean; result?: unknown; error?: string }
  | { type: 'SCREENSHOT_RESULT'; commandId: string; dataUrl?: string; error?: string }
  | { type: 'SNAPSHOT_RESULT'; commandId: string; html?: string; elements?: unknown[]; error?: string };

// Helper to send message from popup to background
export function sendToBackground<T extends PopupToBackgroundMessage>(
  message: T
): Promise<BackgroundToPopupResponse> {
  return chrome.runtime.sendMessage(message);
}

// Helper to send message from background to content script
export function sendToContentScript<T extends BackgroundToContentMessage>(
  tabId: number,
  message: T
): Promise<ContentToBackgroundMessage> {
  return chrome.tabs.sendMessage(tabId, message);
}
