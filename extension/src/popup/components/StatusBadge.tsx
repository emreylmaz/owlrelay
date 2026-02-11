import type { ConnectionState } from '../../shared/types';

interface StatusBadgeProps {
  status: ConnectionState['status'];
}

const statusLabels: Record<ConnectionState['status'], string> = {
  disconnected: 'Disconnected',
  connecting: 'Connecting...',
  connected: 'Connected',
  error: 'Error',
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const isPulsing = status === 'connecting';
  
  return (
    <div className={`status-badge ${status}`}>
      <span className={`status-dot ${isPulsing ? 'pulse' : ''}`} />
      {statusLabels[status]}
    </div>
  );
}
