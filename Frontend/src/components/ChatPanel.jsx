import { LoaderCircle, SendHorizontal, X } from 'lucide-react';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import Avatar from './Avatar';

const MAX_LENGTH = 2000;
// Сообщения одного автора, отправленные подряд в пределах этого окна,
// склеиваются в группу (без повторного аватара и имени).
const GROUP_WINDOW_MS = 5 * 60 * 1000;

function formatTime(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
}

function formatDay(value) {
  const date = new Date(value);
  const today = new Date();
  const yesterday = new Date();
  yesterday.setDate(today.getDate() - 1);

  if (date.toDateString() === today.toDateString()) return 'Сегодня';
  if (date.toDateString() === yesterday.toDateString()) return 'Вчера';
  return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
}

export default function ChatPanel({ chat, user, open, onClose }) {
  const { messages, hasMore, loadingOlder, error, ready, sendMessage, loadOlder } = chat;

  const [draft, setDraft] = useState('');
  const listRef = useRef(null);
  const inputRef = useRef(null);
  // Был ли список прокручен к самому низу до прихода новых сообщений.
  const stickToBottomRef = useRef(true);
  // Высота списка перед подгрузкой старых — чтобы после неё вернуть
  // пользователя на то же место, а не прыгнуть наверх.
  const prevScrollHeightRef = useRef(null);

  // После каждого обновления списка решаем, куда прокрутить.
  useLayoutEffect(() => {
    const list = listRef.current;
    if (!list) return;

    if (prevScrollHeightRef.current !== null) {
      // Подгрузились старые сообщения сверху — сохраняем позицию.
      list.scrollTop += list.scrollHeight - prevScrollHeightRef.current;
      prevScrollHeightRef.current = null;
      return;
    }

    if (stickToBottomRef.current) {
      list.scrollTop = list.scrollHeight;
    }
  }, [messages]);

  // При открытии панели — вниз к последним сообщениям и фокус в поле ввода.
  useEffect(() => {
    if (!open) return;
    const list = listRef.current;
    if (list) list.scrollTop = list.scrollHeight;
    stickToBottomRef.current = true;
    inputRef.current?.focus();
  }, [open]);

  const handleScroll = () => {
    const list = listRef.current;
    if (!list) return;

    const distanceToBottom = list.scrollHeight - list.scrollTop - list.clientHeight;
    stickToBottomRef.current = distanceToBottom < 60;

    if (list.scrollTop < 80 && hasMore && !loadingOlder && messages.length > 0) {
      prevScrollHeightRef.current = list.scrollHeight;
      loadOlder();
    }
  };

  const submit = (event) => {
    event?.preventDefault();
    if (!draft.trim()) return;
    if (sendMessage(draft)) {
      setDraft('');
      stickToBottomRef.current = true;
    }
  };

  const handleKeyDown = (event) => {
    // Enter — отправить, Shift+Enter — новая строка.
    if (event.key === 'Enter' && !event.shiftKey && !event.nativeEvent.isComposing) {
      event.preventDefault();
      submit();
    }
  };

  return (
    <aside className={`chat-panel ${open ? 'is-open' : ''}`} aria-hidden={!open}>
      <header className="chat-header">
        <div>
          <span className="eyebrow">Чат комнаты</span>
          {!ready && <small>Нет соединения</small>}
        </div>
        <button className="chat-close" type="button" onClick={onClose} aria-label="Закрыть чат">
          <X size={18} />
        </button>
      </header>

      <div className="chat-list" ref={listRef} onScroll={handleScroll}>
        {loadingOlder && (
          <div className="chat-loader"><LoaderCircle size={16} className="spin" /> Загружаем историю…</div>
        )}
        {!hasMore && messages.length > 0 && (
          <div className="chat-start">Начало переписки</div>
        )}
        {messages.length === 0 && (
          <div className="chat-empty">
            <strong>Пока тихо</strong>
            <span>Напишите первое сообщение — его увидят все в комнате.</span>
          </div>
        )}

        {messages.map((msg, index) => {
          const prev = messages[index - 1];
          const isOwn = msg.user_id === user.id;
          const newDay = !prev || new Date(prev.created_at).toDateString() !== new Date(msg.created_at).toDateString();
          const grouped =
            !newDay &&
            prev &&
            prev.user_id === msg.user_id &&
            new Date(msg.created_at) - new Date(prev.created_at) < GROUP_WINDOW_MS;

          return (
            <div key={msg.id}>
              {newDay && <div className="chat-day">{formatDay(msg.created_at)}</div>}
              <div className={`chat-message ${isOwn ? 'is-own' : ''} ${grouped ? 'is-grouped' : ''}`}>
                {!isOwn && (
                  <div className="chat-avatar">
                    {!grouped && <Avatar name={msg.user_name || '?'} src={msg.avatar} size="sm" />}
                  </div>
                )}
                <div className="chat-bubble">
                  {!grouped && !isOwn && <span className="chat-author">{msg.user_name || 'Участник'}</span>}
                  <p>{msg.text}</p>
                  <time dateTime={msg.created_at}>{formatTime(msg.created_at)}</time>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <form className="chat-form" onSubmit={submit}>
        {error && <div className="chat-error">{error}</div>}
        <div className="chat-input-row">
          <textarea
            ref={inputRef}
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={ready ? 'Сообщение…' : 'Подключение…'}
            maxLength={MAX_LENGTH}
            rows={1}
            disabled={!ready}
          />
          <button
            className="chat-send"
            type="submit"
            disabled={!ready || !draft.trim()}
            aria-label="Отправить"
          >
            <SendHorizontal size={18} />
          </button>
        </div>
      </form>
    </aside>
  );
}
