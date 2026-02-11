// DOM snapshot for content script

interface InteractiveElement {
  selector: string;
  type: 'button' | 'link' | 'input' | 'select' | 'textarea';
  text?: string;
  placeholder?: string;
  value?: string;
  name?: string;
  id?: string;
}

interface SnapshotResult {
  html: string;
  elements: InteractiveElement[];
  url: string;
  title: string;
  truncated: boolean;
}

// Generate a simplified DOM snapshot
export function captureSnapshot(maxDepth = 10, maxLength = 100000): SnapshotResult {
  const elements: InteractiveElement[] = [];
  let truncated = false;
  
  // Serialize the DOM
  function serializeNode(node: Node, depth: number): string {
    if (depth > maxDepth) {
      return '<!-- truncated -->';
    }
    
    if (node.nodeType === Node.TEXT_NODE) {
      const text = node.textContent?.trim() || '';
      return text ? escapeHtml(text) : '';
    }
    
    if (node.nodeType !== Node.ELEMENT_NODE) {
      return '';
    }
    
    const element = node as Element;
    const tagName = element.tagName.toLowerCase();
    
    // Skip script, style, and hidden elements
    if (tagName === 'script' || tagName === 'style' || tagName === 'noscript') {
      return '';
    }
    
    const style = window.getComputedStyle(element);
    if (style.display === 'none' || style.visibility === 'hidden') {
      return '';
    }
    
    // Build attributes string (only relevant ones)
    const attrs: string[] = [];
    const relevantAttrs = ['id', 'class', 'href', 'src', 'alt', 'title', 'name', 'type', 'placeholder', 'value', 'role', 'aria-label'];
    
    for (const attr of relevantAttrs) {
      const value = element.getAttribute(attr);
      if (value) {
        attrs.push(`${attr}="${escapeAttr(value)}"`);
      }
    }
    
    // Track interactive elements
    const interactive = getInteractiveInfo(element);
    if (interactive) {
      elements.push(interactive);
    }
    
    // Self-closing tags
    const selfClosing = ['img', 'br', 'hr', 'input', 'meta', 'link'];
    if (selfClosing.includes(tagName)) {
      return `<${tagName}${attrs.length ? ' ' + attrs.join(' ') : ''} />`;
    }
    
    // Serialize children
    const children: string[] = [];
    for (const child of element.childNodes) {
      children.push(serializeNode(child, depth + 1));
    }
    
    const childContent = children.filter(c => c).join('');
    
    return `<${tagName}${attrs.length ? ' ' + attrs.join(' ') : ''}>${childContent}</${tagName}>`;
  }
  
  // Get info about interactive elements
  function getInteractiveInfo(element: Element): InteractiveElement | null {
    const tagName = element.tagName.toLowerCase();
    const selector = generateSelector(element);
    
    if (tagName === 'button' || (element as HTMLButtonElement).type === 'button' || (element as HTMLButtonElement).type === 'submit') {
      return {
        selector,
        type: 'button',
        text: element.textContent?.trim().slice(0, 100),
        name: element.getAttribute('name') || undefined,
        id: element.id || undefined,
      };
    }
    
    if (tagName === 'a') {
      return {
        selector,
        type: 'link',
        text: element.textContent?.trim().slice(0, 100),
        id: element.id || undefined,
      };
    }
    
    if (tagName === 'input') {
      const input = element as HTMLInputElement;
      if (['button', 'submit', 'reset'].includes(input.type)) {
        return {
          selector,
          type: 'button',
          text: input.value || undefined,
          name: input.name || undefined,
          id: input.id || undefined,
        };
      }
      return {
        selector,
        type: 'input',
        placeholder: input.placeholder || undefined,
        value: input.type !== 'password' ? input.value : undefined,
        name: input.name || undefined,
        id: input.id || undefined,
      };
    }
    
    if (tagName === 'textarea') {
      const textarea = element as HTMLTextAreaElement;
      return {
        selector,
        type: 'textarea',
        placeholder: textarea.placeholder || undefined,
        value: textarea.value?.slice(0, 100),
        name: textarea.name || undefined,
        id: textarea.id || undefined,
      };
    }
    
    if (tagName === 'select') {
      const select = element as HTMLSelectElement;
      return {
        selector,
        type: 'select',
        value: select.value,
        name: select.name || undefined,
        id: select.id || undefined,
      };
    }
    
    // Check for clickable role
    const role = element.getAttribute('role');
    if (role === 'button' || role === 'link') {
      return {
        selector,
        type: role as 'button' | 'link',
        text: element.textContent?.trim().slice(0, 100),
        id: element.id || undefined,
      };
    }
    
    return null;
  }
  
  // Generate a unique selector for an element
  function generateSelector(element: Element): string {
    if (element.id) {
      return `#${element.id}`;
    }
    
    const tagName = element.tagName.toLowerCase();
    
    // Try unique class combination
    if (element.className && typeof element.className === 'string') {
      const classes = element.className.split(/\s+/).filter(c => c && !c.match(/^(js-|_)/));
      if (classes.length > 0) {
        const selector = `${tagName}.${classes.slice(0, 2).join('.')}`;
        if (document.querySelectorAll(selector).length === 1) {
          return selector;
        }
      }
    }
    
    // Try with attributes
    const name = element.getAttribute('name');
    if (name) {
      const selector = `${tagName}[name="${name}"]`;
      if (document.querySelectorAll(selector).length === 1) {
        return selector;
      }
    }
    
    // Fall back to nth-child
    const parent = element.parentElement;
    if (parent) {
      const siblings = Array.from(parent.children).filter(c => c.tagName === element.tagName);
      const index = siblings.indexOf(element) + 1;
      const parentSelector = parent.tagName.toLowerCase();
      return `${parentSelector} > ${tagName}:nth-child(${index})`;
    }
    
    return tagName;
  }
  
  // Start serialization
  let html = serializeNode(document.body, 0);
  
  // Truncate if too long
  if (html.length > maxLength) {
    html = html.slice(0, maxLength);
    truncated = true;
  }
  
  return {
    html,
    elements,
    url: window.location.href,
    title: document.title,
    truncated,
  };
}

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function escapeAttr(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/"/g, '&quot;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}
