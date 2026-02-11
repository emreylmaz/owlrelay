import { useState, useEffect } from 'preact/hooks';
import { getRelayUrl, getToken } from '../../shared/storage';
import { DEFAULT_RELAY_URL } from '../../shared/constants';

interface ServerConfigProps {
  connected: boolean;
  loading: boolean;
  onConnect: (relayUrl: string, token: string) => void;
  onDisconnect: () => void;
}

export function ServerConfig({ connected, loading, onConnect, onDisconnect }: ServerConfigProps) {
  const [relayUrl, setRelayUrl] = useState(DEFAULT_RELAY_URL);
  const [token, setToken] = useState('');

  // Load saved values
  useEffect(() => {
    Promise.all([getRelayUrl(), getToken()]).then(([savedUrl, savedToken]) => {
      if (savedUrl) setRelayUrl(savedUrl);
      if (savedToken) setToken(savedToken);
    });
  }, []);

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (connected) {
      onDisconnect();
    } else {
      onConnect(relayUrl, token);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="form-group">
        <label className="form-label" htmlFor="relay-url">
          Relay URL
        </label>
        <input
          id="relay-url"
          type="url"
          className="form-input"
          placeholder="ws://localhost:3000"
          value={relayUrl}
          onChange={(e) => setRelayUrl((e.target as HTMLInputElement).value)}
          disabled={connected || loading}
          required
        />
      </div>

      <div className="form-group">
        <label className="form-label" htmlFor="token">
          Token
        </label>
        <input
          id="token"
          type="password"
          className="form-input"
          placeholder="owl_xxxxxxxxxxxxx"
          value={token}
          onChange={(e) => setToken((e.target as HTMLInputElement).value)}
          disabled={connected || loading}
          required
        />
      </div>

      <button
        type="submit"
        className={`btn btn-block ${connected ? 'btn-danger' : 'btn-primary'}`}
        disabled={loading || (!connected && (!relayUrl || !token))}
      >
        {loading ? 'Please wait...' : connected ? 'Disconnect' : 'Connect'}
      </button>
    </form>
  );
}
