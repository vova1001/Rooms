import { ArrowDown, LoaderCircle, SendHorizontal, X } from 'lucide-react';
import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import Avatar from './Avatar';

const MAX_LENGTH = 2000;
// Сообщения одного автора, отправленные подряд в пределах этого окна,
// склеиваются в группу (без повторного аватара и имени).
const GROUP_WINDOW_MS = 5 * 60 * 1000;
// Максимальная высота поля ввода, дальше — прокрутка внутри него.
const INPUT_MAX_HEIGHT = 160;

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

function sameDay(a, b) {
  return new Date(a.created_at).toDateString() === new Date(b.created_at).toDateString();
}

// Продолжает ли сообщение b группу сообщения a.
function continuesGroup(a, b) {
  return Boolean(
    a && b &&
    a.user_id === b.user_id &&
    sameDay(a, b) &&
    new Date(b.created_at) - new Date(a.created_at) < GROUP_WINDOW_MS,
  );
}

function pluralOnline(count) {
  const mod10 = count % 10;
  const mod100 = count % 100;
  if (mod10 === 1 && mod100 !== 11) return `${count} участник в комнате`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return `${count} участника в комнате`;
  return `${count} участников в комнате`;
}

export default function ChatPanel({ chat, user, open, onClose, title = 'Чат комнаты', online = 0 }) {
  const { messages, hasMore, loadingOlder, error, ready, sendMessage, loadOlder } = chat;

  const [draft, setDraft] = useState('');
  const [showScrollDown, setShowScrollDown] = useState(false);
  const listRef = useRef(null);
  const inputRef = useRef(null);
  // Был ли список прокручен к самому низу до прихода новых сообщений.
  const stickToBottomRef = useRef(true);
  // Высота списка перед подгрузкой старых — чтобы после неё вернуть
  // пользователя на то же место, а не прыгнуть наверх.
  const prevScrollHeightRef = useRef(null);

  const scrollToBottom = (smooth = false) => {
    const list = listRef.current;
    if (!list) return;
    list.scrollTo({ top: list.scrollHeight, behavior: smooth ? 'smooth' : 'auto' });
    stickToBottomRef.current = true;
    setShowScrollDown(false);
  };

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
    } else {
      setShowScrollDown(true);
    }
  }, [messages]);

  // При открытии панели — вниз к последним сообщениям и фокус в поле ввода.
  useEffect(() => {
    if (!open) return;
    scrollToBottom();
    inputRef.current?.focus();
  }, [open]);

  // Поле ввода растёт вместе с текстом (field-sizing есть не во всех браузерах).
  useLayoutEffect(() => {
    const input = inputRef.current;
    if (!input) return;
    input.style.height = 'auto';
    input.style.height = `${Math.min(input.scrollHeight, INPUT_MAX_HEIGHT)}px`;
  }, [draft]);

  const handleScroll = () => {
    const list = listRef.current;
    if (!list) return;

    const distanceToBottom = list.scrollHeight - list.scrollTop - list.clientHeight;
    stickToBottomRef.current = distanceToBottom < 60;
    if (stickToBottomRef.current) setShowScrollDown(false);
    else if (distanceToBottom > 300) setShowScrollDown(true);

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
      inputRef.current?.focus();
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
    <aside className={`chat-panel ${open ? 'is-open' : ''}`} aria-hidden={!open} aria-label="Чат комнаты">
      <header className="chat-header">
        <div className="chat-title">
          <strong>{title}</strong>
          <small className={ready ? 'is-online' : 'is-offline'}>
            <span className="chat-status-dot" />
            {ready ? pluralOnline(online) : 'Нет соединения'}
          </small>
        </div>
        <button className="chat-close" type="button" onClick={onClose} aria-label="Закрыть чат">
          <X size={18} />
        </button>
      </header>

      <div className="chat-body">
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
            const next = messages[index + 1];
            const isOwn = msg.user_id === user.id;
            const newDay = !prev || !sameDay(prev, msg);
            const isFirst = !continuesGroup(prev, msg);
            const isLast = !continuesGroup(msg, next);

            return (
              <div key={msg.id} className="chat-row">
                {newDay && <div className="chat-day"><span>{formatDay(msg.created_at)}</span></div>}
                <div
                  className={[
                    'chat-message',
                    isOwn ? 'is-own' : '',
                    isFirst ? 'is-first' : '',
                    isLast ? 'is-last' : '',
                  ].join(' ')}
                >
                  {!isOwn && (
                    <div className="chat-avatar">
                      {isLast && <Avatar name={msg.user_name || '?'} src={msg.avatar} size="sm" />}
                    </div>
                  )}
                  <div className="chat-bubble">
                    {isFirst && !isOwn && <span className="chat-author">{msg.user_name || 'Участник'}</span>}
                    <p>
                      {msg.text}
                      <time dateTime={msg.created_at}>{formatTime(msg.created_at)}</time>
                    </p>
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        {showScrollDown && (
          <button className="chat-scroll-down" type="button" onClick={() => scrollToBottom(true)} aria-label="К последним сообщениям">
            <ArrowDown size={18} />
          </button>
        )}
      </div>

      <form className="chat-form" onSubmit={submit}>
        {error && <div className="chat-error">{error}</div>}
        <div className="chat-input-row">
          <textarea
            ref={inputRef}
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={ready ? 'Написать сообщение…' : 'Подключение…'}
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
        {draft.length > MAX_LENGTH - 200 && (
          <div className="chat-counter">{draft.length} / {MAX_LENGTH}</div>
        )}
      </form>
    </aside>
  );
}
