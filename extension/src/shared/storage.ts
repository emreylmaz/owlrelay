import type { StorageData, AttachedTab } from './types';
import { DEFAULT_RELAY_URL } from './constants';

const STORAGE_KEYS = {
  RELAY_URL: 'relayUrl',
  TOKEN: 'token',
  ATTACHED_TABS: 'attachedTabs',
} as const;

// Get relay URL
export async function getRelayUrl(): Promise<string> {
  const result = await chrome.storage.local.get(STORAGE_KEYS.RELAY_URL);
  return result[STORAGE_KEYS.RELAY_URL] || DEFAULT_RELAY_URL;
}

// Set relay URL
export async function setRelayUrl(url: string): Promise<void> {
  await chrome.storage.local.set({ [STORAGE_KEYS.RELAY_URL]: url });
}

// Get token
export async function getToken(): Promise<string> {
  const result = await chrome.storage.local.get(STORAGE_KEYS.TOKEN);
  return result[STORAGE_KEYS.TOKEN] || '';
}

// Set token
export async function setToken(token: string): Promise<void> {
  await chrome.storage.local.set({ [STORAGE_KEYS.TOKEN]: token });
}

// Get attached tabs
export async function getAttachedTabs(): Promise<AttachedTab[]> {
  const result = await chrome.storage.local.get(STORAGE_KEYS.ATTACHED_TABS);
  return result[STORAGE_KEYS.ATTACHED_TABS] || [];
}

// Set attached tabs
export async function setAttachedTabs(tabs: AttachedTab[]): Promise<void> {
  await chrome.storage.local.set({ [STORAGE_KEYS.ATTACHED_TABS]: tabs });
}

// Add attached tab
export async function addAttachedTab(tab: AttachedTab): Promise<void> {
  const tabs = await getAttachedTabs();
  const existing = tabs.findIndex(t => t.tabId === tab.tabId);
  if (existing >= 0) {
    tabs[existing] = tab;
  } else {
    tabs.push(tab);
  }
  await setAttachedTabs(tabs);
}

// Remove attached tab
export async function removeAttachedTab(tabId: number): Promise<void> {
  const tabs = await getAttachedTabs();
  const filtered = tabs.filter(t => t.tabId !== tabId);
  await setAttachedTabs(filtered);
}

// Get all storage data
export async function getAllStorage(): Promise<StorageData> {
  const result = await chrome.storage.local.get([
    STORAGE_KEYS.RELAY_URL,
    STORAGE_KEYS.TOKEN,
    STORAGE_KEYS.ATTACHED_TABS,
  ]);
  return {
    relayUrl: result[STORAGE_KEYS.RELAY_URL] || DEFAULT_RELAY_URL,
    token: result[STORAGE_KEYS.TOKEN] || '',
    attachedTabs: result[STORAGE_KEYS.ATTACHED_TABS] || [],
  };
}

// Clear all storage
export async function clearStorage(): Promise<void> {
  await chrome.storage.local.clear();
}
