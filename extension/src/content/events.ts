// Event injection for content script
import { findElement, getElementAtPoint, getElementCenter, isInputElement, isContentEditable, focusElement, getScrollableParent } from './dom';
import type { ClickAction, TypeAction, ScrollAction } from '../shared/types';

// Execute click action
export function executeClick(action: ClickAction): { success: boolean; error?: string } {
  let element: Element | null = null;
  let x: number;
  let y: number;
  
  if (action.selector) {
    element = findElement(action.selector);
    if (!element) {
      return { success: false, error: `Element not found: ${action.selector}` };
    }
    const center = getElementCenter(element);
    x = center.x;
    y = center.y;
  } else if (action.coordinates) {
    x = action.coordinates.x;
    y = action.coordinates.y;
    element = getElementAtPoint(x, y);
  } else {
    return { success: false, error: 'No selector or coordinates provided' };
  }
  
  // Convert button name to button number
  const buttonMap: Record<string, number> = {
    left: 0,
    middle: 1,
    right: 2,
  };
  const button = buttonMap[action.button || 'left'];
  
  // Build modifier flags
  const modifiers = action.modifiers || [];
  const ctrlKey = modifiers.includes('ctrl');
  const shiftKey = modifiers.includes('shift');
  const altKey = modifiers.includes('alt');
  const metaKey = modifiers.includes('meta');
  
  // Create and dispatch mouse events
  const eventInit: MouseEventInit = {
    bubbles: true,
    cancelable: true,
    view: window,
    button,
    buttons: 1 << button,
    clientX: x,
    clientY: y,
    screenX: x,
    screenY: y,
    ctrlKey,
    shiftKey,
    altKey,
    metaKey,
  };
  
  const target = element || document.elementFromPoint(x, y) || document.body;
  
  // Dispatch mousedown, mouseup, click sequence
  target.dispatchEvent(new MouseEvent('mousedown', eventInit));
  target.dispatchEvent(new MouseEvent('mouseup', eventInit));
  
  if (button === 0) {
    target.dispatchEvent(new MouseEvent('click', eventInit));
  } else if (button === 2) {
    target.dispatchEvent(new MouseEvent('contextmenu', eventInit));
  }
  
  return { success: true };
}

// Execute type action
export async function executeType(action: TypeAction): Promise<{ success: boolean; error?: string }> {
  const element = findElement(action.selector);
  if (!element) {
    return { success: false, error: `Element not found: ${action.selector}` };
  }
  
  // Focus the element
  focusElement(element);
  
  // Clear if requested
  if (action.clear) {
    if (isInputElement(element)) {
      element.value = '';
      element.dispatchEvent(new Event('input', { bubbles: true }));
    } else if (isContentEditable(element)) {
      (element as HTMLElement).innerHTML = '';
      element.dispatchEvent(new Event('input', { bubbles: true }));
    }
  }
  
  // Type each character
  const delay = action.delay || 0;
  
  for (const char of action.text) {
    // Dispatch keydown, keypress, keyup for each character
    const keyEventInit: KeyboardEventInit = {
      bubbles: true,
      cancelable: true,
      key: char,
      code: `Key${char.toUpperCase()}`,
      charCode: char.charCodeAt(0),
      keyCode: char.charCodeAt(0),
    };
    
    element.dispatchEvent(new KeyboardEvent('keydown', keyEventInit));
    element.dispatchEvent(new KeyboardEvent('keypress', keyEventInit));
    
    // Actually insert the character
    if (isInputElement(element)) {
      const start = element.selectionStart || 0;
      const end = element.selectionEnd || 0;
      const value = element.value;
      element.value = value.slice(0, start) + char + value.slice(end);
      element.selectionStart = element.selectionEnd = start + 1;
    } else if (isContentEditable(element)) {
      document.execCommand('insertText', false, char);
    }
    
    element.dispatchEvent(new Event('input', { bubbles: true }));
    element.dispatchEvent(new KeyboardEvent('keyup', keyEventInit));
    
    // Delay between characters
    if (delay > 0) {
      await new Promise(resolve => setTimeout(resolve, delay));
    }
  }
  
  // Dispatch change event
  element.dispatchEvent(new Event('change', { bubbles: true }));
  
  return { success: true };
}

// Execute scroll action
export function executeScroll(action: ScrollAction): { success: boolean; error?: string } {
  let scrollTarget: Element | Window;
  
  if (action.selector) {
    const element = findElement(action.selector);
    if (!element) {
      return { success: false, error: `Element not found: ${action.selector}` };
    }
    scrollTarget = getScrollableParent(element) || window;
  } else {
    scrollTarget = window;
  }
  
  const amount = action.amount || 100;
  let deltaX = 0;
  let deltaY = 0;
  
  switch (action.direction) {
    case 'up':
      deltaY = -amount;
      break;
    case 'down':
      deltaY = amount;
      break;
    case 'left':
      deltaX = -amount;
      break;
    case 'right':
      deltaX = amount;
      break;
  }
  
  if (scrollTarget === window) {
    window.scrollBy({
      left: deltaX,
      top: deltaY,
      behavior: 'smooth',
    });
  } else {
    (scrollTarget as Element).scrollBy({
      left: deltaX,
      top: deltaY,
      behavior: 'smooth',
    });
  }
  
  return { success: true };
}
