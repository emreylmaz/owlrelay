// DOM utility functions for content script

// Find element by selector
export function findElement(selector: string): Element | null {
  try {
    return document.querySelector(selector);
  } catch {
    return null;
  }
}

// Get element at coordinates
export function getElementAtPoint(x: number, y: number): Element | null {
  return document.elementFromPoint(x, y);
}

// Check if element is visible
export function isElementVisible(element: Element): boolean {
  const rect = element.getBoundingClientRect();
  const style = window.getComputedStyle(element);
  
  return (
    rect.width > 0 &&
    rect.height > 0 &&
    style.visibility !== 'hidden' &&
    style.display !== 'none' &&
    style.opacity !== '0'
  );
}

// Get element center coordinates
export function getElementCenter(element: Element): { x: number; y: number } {
  const rect = element.getBoundingClientRect();
  return {
    x: rect.left + rect.width / 2,
    y: rect.top + rect.height / 2,
  };
}

// Scroll element into view
export function scrollIntoView(element: Element): void {
  element.scrollIntoView({
    behavior: 'smooth',
    block: 'center',
    inline: 'center',
  });
}

// Check if element is input-like
export function isInputElement(element: Element): element is HTMLInputElement | HTMLTextAreaElement {
  const tagName = element.tagName.toLowerCase();
  return tagName === 'input' || tagName === 'textarea';
}

// Check if element is contenteditable
export function isContentEditable(element: Element): boolean {
  return (element as HTMLElement).isContentEditable === true;
}

// Focus element
export function focusElement(element: Element): void {
  if (element instanceof HTMLElement) {
    element.focus();
  }
}

// Get scrollable parent
export function getScrollableParent(element: Element): Element | null {
  let parent: Element | null = element.parentElement;
  
  while (parent) {
    const style = window.getComputedStyle(parent);
    const overflow = style.overflow + style.overflowX + style.overflowY;
    
    if (/(auto|scroll)/.test(overflow)) {
      return parent;
    }
    parent = parent.parentElement;
  }
  
  return document.documentElement;
}
