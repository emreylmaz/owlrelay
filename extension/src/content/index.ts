// OwlRelay Content Script
import type { CommandAction } from '../shared/types';
import type { BackgroundToContentMessage, ContentToBackgroundMessage } from '../shared/messages';
import { executeClick, executeType, executeScroll } from './events';
import { captureSnapshot } from './snapshot';

console.log('[OwlRelay] Content script loaded');

// Listen for messages from background
chrome.runtime.onMessage.addListener((
  message: BackgroundToContentMessage,
  _sender,
  sendResponse: (response: ContentToBackgroundMessage) => void
) => {
  handleMessage(message).then(sendResponse);
  return true; // Keep channel open for async response
});

async function handleMessage(message: BackgroundToContentMessage): Promise<ContentToBackgroundMessage> {
  switch (message.type) {
    case 'EXECUTE_COMMAND': {
      return executeCommand(message.commandId, message.action as CommandAction);
    }
    
    case 'GET_SNAPSHOT': {
      try {
        const result = captureSnapshot(message.maxDepth, message.maxLength);
        return {
          type: 'SNAPSHOT_RESULT',
          commandId: message.commandId,
          html: result.html,
          elements: result.elements,
        };
      } catch (err) {
        return {
          type: 'SNAPSHOT_RESULT',
          commandId: message.commandId,
          error: err instanceof Error ? err.message : 'Failed to capture snapshot',
        };
      }
    }
    
    case 'TAKE_SCREENSHOT': {
      // Screenshot is handled by background script via chrome.tabs.captureVisibleTab
      // This message type is here for completeness but shouldn't be called
      return {
        type: 'SCREENSHOT_RESULT',
        commandId: message.commandId,
        error: 'Screenshot should be handled by background script',
      };
    }
  }
}

async function executeCommand(
  commandId: string,
  action: CommandAction
): Promise<ContentToBackgroundMessage> {
  try {
    switch (action.kind) {
      case 'click': {
        const result = executeClick(action);
        return {
          type: 'COMMAND_RESULT',
          commandId,
          success: result.success,
          error: result.error,
        };
      }
      
      case 'type': {
        const result = await executeType(action);
        return {
          type: 'COMMAND_RESULT',
          commandId,
          success: result.success,
          error: result.error,
        };
      }
      
      case 'scroll': {
        const result = executeScroll(action);
        return {
          type: 'COMMAND_RESULT',
          commandId,
          success: result.success,
          error: result.error,
        };
      }
      
      case 'navigate': {
        // Navigate to URL
        window.location.href = action.url;
        return {
          type: 'COMMAND_RESULT',
          commandId,
          success: true,
        };
      }
      
      default: {
        return {
          type: 'COMMAND_RESULT',
          commandId,
          success: false,
          error: `Unknown action kind: ${(action as { kind: string }).kind}`,
        };
      }
    }
  } catch (err) {
    return {
      type: 'COMMAND_RESULT',
      commandId,
      success: false,
      error: err instanceof Error ? err.message : 'Command execution failed',
    };
  }
}
