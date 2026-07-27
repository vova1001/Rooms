import { useEffect, useState } from 'react';

const GATEWAY_WS = import.meta.env.VITE_GATEWAY_WS || 'ws://localhost:8081/ws';

export function useGatewayJoin({ roomId, user, enabled = true }) {
  const [connection, setConnection] = useState(null);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  useEffect(() => {
    if (!enabled || !roomId || !user?.id) return undefined;

    setStatus('connecting');
    setError('');

    const url = new URL(GATEWAY_WS);
    url.searchParams.set('room_id', roomId);
    url.searchParams.set('user_id', user.id);
    url.searchParams.set('user_name', user.username || 'User');

    const socket = new WebSocket(url.toString());

    socket.onopen = () => setStatus('waiting');
    socket.onerror = () => {
      setError('Не удалось подключиться к Gateway');
      setStatus('error');
    };
    socket.onclose = () => {
      setStatus((current) => (current === 'error' ? current : 'closed'));
    };
    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        if (message.type === 'error') {
          throw new Error(message.message || message.error || 'Gateway вернул ошибку');
        }
        if (message.type === 'joined') {
          const data = message.data || message;
          const livekit = data.livekit || data.ConnectionData || data.connection_data;
          if (!livekit?.url || !livekit?.token) {
            throw new Error('Gateway не вернул LiveKit URL или token');
          }
          setConnection({ url: livekit.url, token: livekit.token });
          setStatus('joined');
        }
      } catch (err) {
        setError(err.message || 'Некорректный ответ Gateway');
        setStatus('error');
      }
    };

    return () => {
      socket.close(1000, 'page leave');
    };
  }, [enabled, roomId, user?.id, user?.username]);

  return { connection, status, error };
}
