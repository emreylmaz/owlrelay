// OwlRelay Background Service Worker
import type { PopupToBackgroundMessage, BackgroundToPopupResponse } from '../shared/messages';
import { connect, disconnect, getConnectionState, onConnectionStateChange } from './websocket';
import { initTabs, attachTab, detachTab, getAttachedTabsForRelay, handleTabUpdate, handleTabRemove } from './tabs';
import { getRelayUrl, getToken, setRelayUrl, setToken } from '../shared/storage';

console.log('[OwlRelay] Background service worker started');

// Initialize on startup
async function init(): Promise<void> {
  await initTabs();
  
  // Auto-connect if we have saved credentials
  const [relayUrl, token] = await Promise.all([getRelayUrl(), getToken()]);
  if (relayUrl && token) {
    console.log('[OwlRelay] Auto-connecting with saved credentials');
    connect(relayUrl, token);
  }
}

init();

// Listen for messages from popup
chrome.runtime.onMessage.addListener((
  message: PopupToBackgroundMessage,
  _sender,
  sendResponse: (response: BackgroundToPopupResponse) => void
) => {
  handlePopupMessage(message).then(sendResponse);
  return true; // Keep channel open for async response
});

async function handlePopupMessage(message: PopupToBackgroundMessage): Promise<BackgroundToPopupResponse> {
  switch (message.type) {
    case 'GET_STATUS': {
      const state = getConnectionState();
      const tabs = getAttachedTabsForRelay();
      return { type: 'STATUS', state, attachedTabs: tabs };
    }
    
    case 'CONNECT': {
      try {
        // Save credentials
        await setRelayUrl(message.relayUrl);
        await setToken(message.token);
        
        // Connect
        connect(message.relayUrl, message.token);
        
        // Wait for connection result
        return new Promise((resolve) => {
          const timeout = setTimeout(() => {
            resolve({ type: 'ERROR', message: 'Connection timeout' });
          }, 10000);
          
          const unsubscribe = onConnectionStateChange((state) => {
            if (state.status === 'connected') {
              clearTimeout(timeout);
              unsubscribe();
              resolve({ type: 'CONNECTED', sessionId: state.sessionId || '' });
            } else if (state.status === 'error') {
              clearTimeout(timeout);
              unsubscribe();
              resolve({ type: 'ERROR', message: state.error || 'Connection failed' });
            }
          });
        });
      } catch (err) {
        return { type: 'ERROR', message: err instanceof Error ? err.message : 'Connection failed' };
      }
    }
    
    case 'DISCONNECT': {
      disconnect();
      // Clear saved token (keep relay URL)
      await setToken('');
      return { type: 'DISCONNECTED' };
    }
    
    case 'ATTACH_TAB': {
      try {
        const tab = await attachTab(message.tabId);
        if (tab) {
          return { type: 'TAB_ATTACHED', tab };
        }
        return { type: 'ERROR', message: 'Failed to attach tab' };
      } catch (err) {
        return { type: 'ERROR', message: err instanceof Error ? err.message : 'Failed to attach tab' };
      }
    }
    
    case 'DETACH_TAB': {
      await detachTab(message.tabId);
      return { type: 'TAB_DETACHED', tabId: message.tabId };
    }
    
    case 'GET_ATTACHED_TABS': {
      const tabs = getAttachedTabsForRelay();
      return { type: 'ATTACHED_TABS', tabs };
    }
  }
}

// Listen for tab updates
chrome.tabs.onUpdated.addListener((tabId, changeInfo) => {
  handleTabUpdate(tabId, changeInfo);
});

// Listen for tab removal
chrome.tabs.onRemoved.addListener((tabId) => {
  handleTabRemove(tabId);
});

// Handle extension install/update
chrome.runtime.onInstalled.addListener((details) => {
  console.log('[OwlRelay] Extension installed/updated:', details.reason);
  
  if (details.reason === 'install') {
    // Open options page on first install
    // chrome.runtime.openOptionsPage();
  }
});

// Keep service worker alive (Manifest V3 workaround)
// The service worker will be woken up by alarms and messages
chrome.alarms.create('keepalive', { periodInMinutes: 0.5 });
chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === 'keepalive') {
    // Just a ping to keep the service worker active
    console.log('[OwlRelay] Keepalive ping');
  }
});
