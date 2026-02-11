// Default relay URL (localhost for development)
export const DEFAULT_RELAY_URL = 'ws://localhost:3000';

// Heartbeat interval in milliseconds
export const HEARTBEAT_INTERVAL = 30_000;

// Reconnect settings
export const RECONNECT_DELAY_BASE = 1000;
export const RECONNECT_DELAY_MAX = 30_000;
export const MAX_RECONNECT_ATTEMPTS = 10;

// Command timeout
export const DEFAULT_COMMAND_TIMEOUT = 10_000;

// Banking and sensitive sites blacklist
export const BLACKLISTED_PATTERNS = [
  // Banking - US
  '*://*.bankofamerica.com/*',
  '*://*.chase.com/*',
  '*://*.wellsfargo.com/*',
  '*://*.citi.com/*',
  '*://*.usbank.com/*',
  
  // Banking - Turkey
  '*://*.garanti.com.tr/*',
  '*://*.isbank.com.tr/*',
  '*://*.yapikredi.com.tr/*',
  '*://*.akbank.com/*',
  '*://*.ziraat.com.tr/*',
  '*://*.qnb.com.tr/*',
  '*://*.halkbank.com.tr/*',
  '*://*.vakifbank.com.tr/*',
  '*://*.denizbank.com/*',
  '*://*.ingbank.com.tr/*',
  
  // Banking - Europe
  '*://*.hsbc.com/*',
  '*://*.barclays.co.uk/*',
  '*://*.lloydsbank.com/*',
  '*://*.natwest.com/*',
  '*://*.deutschebank.de/*',
  '*://*.bnpparibas.com/*',
  
  // Password managers
  '*://*.1password.com/*',
  '*://*.lastpass.com/*',
  '*://*.bitwarden.com/*',
  '*://*.dashlane.com/*',
  '*://*.keeper.io/*',
  
  // Crypto exchanges
  '*://*.coinbase.com/*',
  '*://*.binance.com/*',
  '*://*.kraken.com/*',
  '*://*.gemini.com/*',
  '*://*.crypto.com/*',
  '*://*.kucoin.com/*',
  
  // Auth pages
  '*://accounts.google.com/*',
  '*://login.microsoftonline.com/*',
  '*://appleid.apple.com/*',
  '*://id.apple.com/*',
  '*://login.live.com/*',
  '*://auth0.com/*',
  '*://login.yahoo.com/*',
  
  // Payment processors
  '*://*.paypal.com/*',
  '*://*.stripe.com/*',
  '*://*.venmo.com/*',
  '*://*.square.com/*',
];

// Check if a URL matches the blacklist
export function isBlacklisted(url: string): boolean {
  try {
    const urlObj = new URL(url);
    const hostname = urlObj.hostname.toLowerCase();
    
    for (const pattern of BLACKLISTED_PATTERNS) {
      // Extract domain from pattern
      const match = pattern.match(/\*:\/\/\*?\.?([^/]+)/);
      if (match) {
        const domain = match[1].replace('*', '');
        if (hostname.includes(domain) || hostname === domain) {
          return true;
        }
      }
    }
    return false;
  } catch {
    return false;
  }
}
