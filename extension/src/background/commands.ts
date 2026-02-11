import type { CommandRequest, CommandResponse, CommandAction } from '../shared/types';
import type { BackgroundToContentMessage, ContentToBackgroundMessage } from '../shared/messages';
import { sendMessage } from './websocket';
import { getAttachedTabByUuid } from './tabs';
import { DEFAULT_COMMAND_TIMEOUT } from '../shared/constants';

// Handle incoming command from relay
export async function handleRelayMessage(command: CommandRequest): Promise<void> {
  const startTime = Date.now();
  
  // Find the attached tab
  const attachedTab = getAttachedTabByUuid(command.tabId);
  if (!attachedTab) {
    sendCommandResponse(command.id, false, startTime, undefined, {
      code: 'TAB_NOT_FOUND',
      message: `Tab ${command.tabId} is not attached`,
    });
    return;
  }
  
  const timeout = command.timeout || DEFAULT_COMMAND_TIMEOUT;
  
  try {
    const result = await executeCommand(attachedTab.tabId, command.id, command.action, timeout);
    sendCommandResponse(command.id, true, startTime, result);
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Unknown error';
    sendCommandResponse(command.id, false, startTime, undefined, {
      code: 'EXECUTION_ERROR',
      message: errorMessage,
    });
  }
}

async function executeCommand(
  tabId: number,
  commandId: string,
  action: CommandAction,
  timeout: number
): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const timer = setTimeout(() => {
      reject(new Error('Command timed out'));
    }, timeout);
    
    // Determine message type based on action
    let message: BackgroundToContentMessage;
    
    if (action.kind === 'screenshot') {
      // Screenshot uses chrome.tabs.captureVisibleTab, handled in background
      clearTimeout(timer);
      captureScreenshot(tabId).then(resolve).catch(reject);
      return;
    } else if (action.kind === 'snapshot') {
      message = {
        type: 'GET_SNAPSHOT',
        commandId,
        maxDepth: action.maxDepth,
        maxLength: action.maxLength,
      };
    } else {
      message = {
        type: 'EXECUTE_COMMAND',
        commandId,
        action,
      };
    }
    
    // Send to content script
    chrome.tabs.sendMessage(tabId, message, (response: ContentToBackgroundMessage) => {
      clearTimeout(timer);
      
      if (chrome.runtime.lastError) {
        reject(new Error(chrome.runtime.lastError.message || 'Failed to communicate with tab'));
        return;
      }
      
      if (!response) {
        reject(new Error('No response from content script'));
        return;
      }
      
      if (response.type === 'COMMAND_RESULT') {
        if (response.success) {
          resolve(response.result);
        } else {
          reject(new Error(response.error || 'Command failed'));
        }
      } else if (response.type === 'SNAPSHOT_RESULT') {
        if (response.error) {
          reject(new Error(response.error));
        } else {
          resolve({ html: response.html, elements: response.elements });
        }
      } else {
        reject(new Error('Unexpected response type'));
      }
    });
  });
}

async function captureScreenshot(tabId: number): Promise<{ data: string; width: number; height: number }> {
  // First, make sure the tab is active
  const tab = await chrome.tabs.get(tabId);
  if (!tab.windowId) {
    throw new Error('Tab has no window');
  }
  
  // Focus the window and tab
  await chrome.windows.update(tab.windowId, { focused: true });
  await chrome.tabs.update(tabId, { active: true });
  
  // Small delay to ensure rendering
  await new Promise(resolve => setTimeout(resolve, 100));
  
  // Capture
  const dataUrl = await chrome.tabs.captureVisibleTab(tab.windowId, {
    format: 'png',
    quality: 90,
  });
  
  // Extract base64 data without the data URL prefix
  const base64Data = dataUrl.split(',')[1];
  
  // Get image dimensions using offscreen canvas (service worker compatible)
  const response = await fetch(dataUrl);
  const blob = await response.blob();
  const bitmap = await createImageBitmap(blob);
  const width = bitmap.width;
  const height = bitmap.height;
  bitmap.close();
  
  return { data: base64Data, width, height };
}

function sendCommandResponse(
  id: string,
  success: boolean,
  startTime: number,
  result?: unknown,
  error?: { code: string; message: string }
): void {
  const response: CommandResponse = {
    type: 'command_response',
    id,
    success,
    result,
    error,
    timing: {
      received: startTime,
      completed: Date.now(),
    },
  };
  
  sendMessage(response);
}
