import { useCallback, useEffect, useRef, useState } from 'react';

const GATEWAY_WS = import.meta.env.VITE_GATEWAY_WS || 'wss://ws.hushh.site/ws';

// Должно совпадать с historyLimit на бэкенде (service/message/service.go).
// Если пришло меньше — значит, более старых сообщений нет.
const HISTORY_PAGE_SIZE = 50;

// Объединяет два списка сообщений: убирает дубли по id и сортирует
// по времени создания (от старых к новым), чтобы у всех был одинаковый порядок.
function mergeMessages(prev, incoming) {
  const byId = new Map();
  for (const msg of prev) byId.set(msg.id, msg);
  for (const msg of incoming) byId.set(msg.id, msg);
  return [...byId.values()].sort(
    (a, b) => new Date(a.created_at) - new Date(b.created_at),
  );
}

export function useGatewayJoin({ roomId, user, enabled = true }) {
  const [connection, setConnection] = useState(null);
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState('');

  // --- чат ---
  const [messages, setMessages] = useState([]);
  const [hasMore, setHasMore] = useState(true);
  const [loadingOlder, setLoadingOlder] = useState(false);
  const [chatError, setChatError] = useState('');
  const [socketOpen, setSocketOpen] = useState(false);
  const socketRef = useRef(null);
  // Актуальный список сообщений для loadOlder без пересоздания функции.
  const messagesRef = useRef(messages);
  messagesRef.current = messages;

  useEffect(() => {
    if (!enabled || !roomId || !user?.id) return undefined;

    setStatus('connecting');
    setError('');
    setMessages([]);
    setHasMore(true);
    setLoadingOlder(false);
    setChatError('');

    const url = new URL(GATEWAY_WS);
    url.searchParams.set('room_id', roomId);
    url.searchParams.set('user_id', user.id);
    url.searchParams.set('user_name', user.username || 'User');

    const socket = new WebSocket(url.toString());
    socketRef.current = socket;

    socket.onopen = () => {
      setStatus('waiting');
      setSocketOpen(true);
    };
    socket.onerror = () => {
      setError('Не удалось подключиться к Gateway');
      setStatus('error');
    };
    socket.onclose = () => {
      // Старый сокет (после смены комнаты) не должен трогать состояние нового.
      if (socketRef.current === socket) setSocketOpen(false);
      setStatus((current) => (current === 'error' ? current : 'closed'));
    };
    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);

        // --- события чата ---
        if (message.type === 'message.created') {
          setMessages((prev) => mergeMessages(prev, [message.data]));
          return;
        }
        if (message.type === 'message.history') {
          const page = message.data || [];
          setMessages((prev) => mergeMessages(prev, page));
          setHasMore(page.length >= HISTORY_PAGE_SIZE);
          setLoadingOlder(false);
          return;
        }
        if (message.type === 'message.error') {
          console.warn('chat error:', message.error);
          setChatError(message.error || 'Ошибка чата');
          setLoadingOlder(false);
          return;
        }

        // --- служебные события комнаты ---
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
      socketRef.current = null;
      setSocketOpen(false);
      socket.close(1000, 'page leave');
    };
  }, [enabled, roomId, user?.id, user?.username]);

  // Отправляет событие по сокету. Возвращает false, если сокет не открыт.
  const sendEvent = useCallback((type, data) => {
    const socket = socketRef.current;
    if (!socket || socket.readyState !== WebSocket.OPEN) return false;
    socket.send(JSON.stringify({ type, data }));
    return true;
  }, []);

  const sendMessage = useCallback((text) => {
    const trimmed = text.trim();
    if (!trimmed) return false;
    setChatError('');
    return sendEvent('message.create', { text: trimmed });
  }, [sendEvent]);

  // Подгружает следующую страницу сообщений, старше самого старого из загруженных.
  const loadOlder = useCallback(() => {
    const oldest = messagesRef.current[0];
    if (!oldest || loadingOlder || !hasMore) return;
    if (sendEvent('message.history', { before: oldest.created_at })) {
      setLoadingOlder(true);
    }
  }, [sendEvent, loadingOlder, hasMore]);

  const chat = {
    messages,
    hasMore,
    loadingOlder,
    error: chatError,
    ready: socketOpen,
    sendMessage,
    loadOlder,
  };

  return { connection, status, error, chat };
}
