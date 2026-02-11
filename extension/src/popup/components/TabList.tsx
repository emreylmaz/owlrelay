import type { AttachedTab } from '../../shared/types';

interface TabListProps {
  tabs: AttachedTab[];
  currentTabId?: number;
  onDetach: (tabId: number) => void;
}

export function TabList({ tabs, currentTabId, onDetach }: TabListProps) {
  // Filter out current tab from list (it's shown separately)
  const otherTabs = tabs.filter(t => t.tabId !== currentTabId);
  
  if (otherTabs.length === 0) {
    return null;
  }

  return (
    <section className="tabs-section">
      <div className="tabs-header">
        <span className="tabs-title">Attached Tabs</span>
        <span className="tabs-count">{otherTabs.length}</span>
      </div>
      
      <ul className="tab-list">
        {otherTabs.map((tab) => (
          <TabItem key={tab.tabId} tab={tab} onDetach={onDetach} />
        ))}
      </ul>
    </section>
  );
}

interface TabItemProps {
  tab: AttachedTab;
  onDetach: (tabId: number) => void;
}

function TabItem({ tab, onDetach }: TabItemProps) {
  const hostname = (() => {
    try {
      return new URL(tab.url).hostname;
    } catch {
      return tab.url;
    }
  })();

  return (
    <li className="tab-item">
      {tab.favIconUrl ? (
        <img src={tab.favIconUrl} alt="" className="tab-favicon" />
      ) : (
        <div className="tab-favicon" style={{ background: '#e5e7eb' }} />
      )}
      <div className="tab-info">
        <div className="tab-title">{tab.title}</div>
        <div className="tab-url">{hostname}</div>
      </div>
      <button
        className="tab-detach"
        onClick={() => onDetach(tab.tabId)}
        title="Detach tab"
      >
        ✕
      </button>
    </li>
  );
}
