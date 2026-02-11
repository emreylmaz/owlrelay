import { useState, useEffect, useCallback } from 'preact/hooks';
import type { ConnectionState, AttachedTab } from '../shared/types';
import type { BackgroundToPopupResponse } from '../shared/messages';
import { ServerConfig } from './components/ServerConfig';
import { StatusBadge } from './components/StatusBadge';
import { TabList } from './components/TabList';
import { CurrentTab } from './components/CurrentTab';

export function App() {
  const [state, setState] = useState<ConnectionState>({ status: 'disconnected' });
  const [attachedTabs, setAttachedTabs] = useState<AttachedTab[]>([]);
  const [currentTab, setCurrentTab] = useState<chrome.tabs.Tab | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  // Fetch initial status
  useEffect(() => {
    fetchStatus();
    getCurrentTab();
  }, []);

  const fetchStatus = useCallback(async () => {
    try {
      const response: BackgroundToPopupResponse = await chrome.runtime.sendMessage({ type: 'GET_STATUS' });
      if (response.type === 'STATUS') {
        setState(response.state);
        setAttachedTabs(response.attachedTabs);
      }
    } catch (err) {
      console.error('Failed to fetch status:', err);
    }
  }, []);

  const getCurrentTab = useCallback(async () => {
    try {
      const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
      setCurrentTab(tab || null);
    } catch (err) {
      console.error('Failed to get current tab:', err);
    }
  }, []);

  const handleConnect = useCallback(async (relayUrl: string, token: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response: BackgroundToPopupResponse = await chrome.runtime.sendMessage({
        type: 'CONNECT',
        relayUrl,
        token,
      });
      
      if (response.type === 'CONNECTED') {
        setState({ status: 'connected', sessionId: response.sessionId });
      } else if (response.type === 'ERROR') {
        setError(response.message);
        setState({ status: 'error', error: response.message });
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Connection failed';
      setError(message);
      setState({ status: 'error', error: message });
    } finally {
      setLoading(false);
    }
  }, []);

  const handleDisconnect = useCallback(async () => {
    setLoading(true);
    try {
      await chrome.runtime.sendMessage({ type: 'DISCONNECT' });
      setState({ status: 'disconnected' });
      setError(null);
    } catch (err) {
      console.error('Failed to disconnect:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleAttachTab = useCallback(async () => {
    if (!currentTab?.id) return;
    
    setError(null);
    try {
      const response: BackgroundToPopupResponse = await chrome.runtime.sendMessage({
        type: 'ATTACH_TAB',
        tabId: currentTab.id,
      });
      
      if (response.type === 'TAB_ATTACHED') {
        setAttachedTabs(prev => [...prev.filter(t => t.tabId !== response.tab.tabId), response.tab]);
      } else if (response.type === 'ERROR') {
        setError(response.message);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to attach tab');
    }
  }, [currentTab]);

  const handleDetachTab = useCallback(async (tabId: number) => {
    try {
      await chrome.runtime.sendMessage({ type: 'DETACH_TAB', tabId });
      setAttachedTabs(prev => prev.filter(t => t.tabId !== tabId));
    } catch (err) {
      console.error('Failed to detach tab:', err);
    }
  }, []);

  const isConnected = state.status === 'connected';
  const isCurrentTabAttached = currentTab?.id ? attachedTabs.some(t => t.tabId === currentTab.id) : false;

  return (
    <div>
      <header className="header">
        <img src="../assets/icon-48.png" alt="" className="header-icon" />
        <h1 className="header-title">OwlRelay</h1>
        <StatusBadge status={state.status} />
      </header>

      {error && (
        <div className="error-message">
          {error}
        </div>
      )}

      <ServerConfig
        connected={isConnected}
        loading={loading}
        onConnect={handleConnect}
        onDisconnect={handleDisconnect}
      />

      {isConnected && (
        <>
          <CurrentTab
            tab={currentTab}
            isAttached={isCurrentTabAttached}
            onAttach={handleAttachTab}
            onDetach={() => currentTab?.id && handleDetachTab(currentTab.id)}
          />

          <TabList
            tabs={attachedTabs}
            currentTabId={currentTab?.id}
            onDetach={handleDetachTab}
          />
        </>
      )}

      <footer className="footer">
        <a href="https://github.com/owlrelay/owlrelay" target="_blank" rel="noopener" className="footer-link">
          GitHub
        </a>
      </footer>
    </div>
  );
}
