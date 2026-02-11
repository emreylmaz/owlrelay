import type { AttachedTab } from '../shared/types';
import { getAttachedTabs, addAttachedTab, removeAttachedTab, setAttachedTabs } from '../shared/storage';
import { isBlacklisted } from '../shared/constants';
import { sendMessage, isConnected } from './websocket';

// In-memory cache of attached tabs
let attachedTabs: AttachedTab[] = [];

// Generate UUID for tab
function generateTabUuid(): string {
  return crypto.randomUUID();
}

// Initialize tabs from storage
export async function initTabs(): Promise<void> {
  attachedTabs = await getAttachedTabs();
  
  // Validate tabs still exist
  const validTabs: AttachedTab[] = [];
  for (const tab of attachedTabs) {
    try {
      const chromeTab = await chrome.tabs.get(tab.tabId);
      if (chromeTab) {
        validTabs.push({
          ...tab,
          url: chromeTab.url || tab.url,
          title: chromeTab.title || tab.title,
          favIconUrl: chromeTab.favIconUrl || tab.favIconUrl,
        });
      }
    } catch {
      // Tab no longer exists
      console.log(`[OwlRelay] Tab ${tab.tabId} no longer exists, removing`);
    }
  }
  
  attachedTabs = validTabs;
  await setAttachedTabs(validTabs);
}

// Attach a tab
export async function attachTab(tabId: number): Promise<AttachedTab | null> {
  // Get tab info
  let chromeTab: chrome.tabs.Tab;
  try {
    chromeTab = await chrome.tabs.get(tabId);
  } catch {
    throw new Error('Tab not found');
  }
  
  const url = chromeTab.url || '';
  
  // Check blacklist
  if (isBlacklisted(url)) {
    throw new Error('This site is on the security blacklist and cannot be controlled');
  }
  
  // Check if already attached
  const existing = attachedTabs.find(t => t.tabId === tabId);
  if (existing) {
    return existing;
  }
  
  // Create new attached tab
  const tab: AttachedTab = {
    tabId,
    uuid: generateTabUuid(),
    url,
    title: chromeTab.title || 'Untitled',
    favIconUrl: chromeTab.favIconUrl,
    attachedAt: Date.now(),
  };
  
  attachedTabs.push(tab);
  await addAttachedTab(tab);
  
  // Update badge for this tab
  await updateTabBadge(tabId, true);
  
  // Notify relay
  if (isConnected()) {
    sendMessage({
      type: 'tab_attach',
      tabId: tab.uuid,
      url: tab.url,
      title: tab.title,
      favIconUrl: tab.favIconUrl,
    });
  }
  
  return tab;
}

// Detach a tab
export async function detachTab(tabId: number): Promise<void> {
  const index = attachedTabs.findIndex(t => t.tabId === tabId);
  if (index < 0) return;
  
  const tab = attachedTabs[index];
  attachedTabs.splice(index, 1);
  await removeAttachedTab(tabId);
  
  // Update badge for this tab
  await updateTabBadge(tabId, false);
  
  // Notify relay
  if (isConnected()) {
    sendMessage({
      type: 'tab_detach',
      tabId: tab.uuid,
    });
  }
}

// Get attached tabs for relay (uses uuid)
export function getAttachedTabsForRelay(): AttachedTab[] {
  return [...attachedTabs];
}

// Get attached tab by chrome tab ID
export function getAttachedTabById(tabId: number): AttachedTab | undefined {
  return attachedTabs.find(t => t.tabId === tabId);
}

// Get attached tab by UUID (for relay messages)
export function getAttachedTabByUuid(uuid: string): AttachedTab | undefined {
  return attachedTabs.find(t => t.uuid === uuid);
}

// Check if tab is attached
export function isTabAttached(tabId: number): boolean {
  return attachedTabs.some(t => t.tabId === tabId);
}

// Update badge for a specific tab
async function updateTabBadge(tabId: number, attached: boolean): Promise<void> {
  try {
    if (attached) {
      await chrome.action.setBadgeText({ tabId, text: '●' });
      await chrome.action.setBadgeBackgroundColor({ tabId, color: '#22c55e' });
    } else {
      await chrome.action.setBadgeText({ tabId, text: '' });
    }
  } catch {
    // Tab might be closed
  }
}

// Handle tab updates from Chrome
export function handleTabUpdate(tabId: number, changeInfo: chrome.tabs.TabChangeInfo): void {
  const tab = attachedTabs.find(t => t.tabId === tabId);
  if (!tab) return;
  
  let updated = false;
  
  if (changeInfo.url) {
    // Check if new URL is blacklisted
    if (isBlacklisted(changeInfo.url)) {
      console.log(`[OwlRelay] Tab ${tabId} navigated to blacklisted site, detaching`);
      detachTab(tabId);
      return;
    }
    tab.url = changeInfo.url;
    updated = true;
  }
  
  if (changeInfo.title) {
    tab.title = changeInfo.title;
    updated = true;
  }
  
  if (changeInfo.favIconUrl) {
    tab.favIconUrl = changeInfo.favIconUrl;
    updated = true;
  }
  
  if (updated && isConnected()) {
    sendMessage({
      type: 'tab_update',
      tabId: tab.uuid,
      url: tab.url,
      title: tab.title,
    });
  }
}

// Handle tab removal from Chrome
export function handleTabRemove(tabId: number): void {
  const index = attachedTabs.findIndex(t => t.tabId === tabId);
  if (index >= 0) {
    const tab = attachedTabs[index];
    attachedTabs.splice(index, 1);
    removeAttachedTab(tabId);
    
    if (isConnected()) {
      sendMessage({
        type: 'tab_detach',
        tabId: tab.uuid,
      });
    }
  }
}
