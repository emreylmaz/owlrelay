interface CurrentTabProps {
  tab: chrome.tabs.Tab | null;
  isAttached: boolean;
  onAttach: () => void;
  onDetach: () => void;
}

export function CurrentTab({ tab, isAttached, onAttach, onDetach }: CurrentTabProps) {
  if (!tab) {
    return null;
  }

  const hostname = (() => {
    try {
      return tab.url ? new URL(tab.url).hostname : '';
    } catch {
      return '';
    }
  })();

  return (
    <div className="current-tab">
      <div className="current-tab-info">
        {tab.favIconUrl ? (
          <img src={tab.favIconUrl} alt="" className="current-tab-icon" />
        ) : (
          <div className="current-tab-icon" style={{ background: '#e5e7eb', borderRadius: 2 }} />
        )}
        <span className="current-tab-title" title={tab.title}>
          {tab.title || hostname || 'Current Tab'}
        </span>
      </div>
      
      {isAttached ? (
        <button className="btn btn-secondary btn-block" onClick={onDetach}>
          Detach This Tab
        </button>
      ) : (
        <button className="btn btn-primary btn-block" onClick={onAttach}>
          Attach This Tab
        </button>
      )}
    </div>
  );
}
