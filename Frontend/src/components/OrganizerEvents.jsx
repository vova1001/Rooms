import { useCallback, useEffect, useRef, useState } from 'react';
import { CalendarDays, Check, ChevronDown, History, RefreshCw, X } from 'lucide-react';
import { decideBooking, getMyEvents } from '../api/http';
import useProfileEventList from '../hooks/useProfileEventList';
import Avatar from './Avatar';
import { groupOrganizerEvents } from '../utils/bookings';

const labels = { pending: 'Ожидает решения', confirmed: 'Принята', rejected: 'Отклонена' };
const errors = {
  'booking already decided': 'Заявка уже рассмотрена. Список обновлён.',
  'event already started': 'Мероприятие уже началось, изменить заявку нельзя.',
  'no seats left': 'Свободных мест больше нет.',
  'booking not found': 'Заявка не найдена или у вас нет прав на неё.',
};
const format = (value) => new Date(value).toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });

export default function OrganizerEvents({ refreshSession, roomId }) {
  const load = useCallback(() => getMyEvents(roomId), [roomId]);
  const { items: events, loading, error, reload, updateItems } = useProfileEventList(load, refreshSession);
  const [pending, setPending] = useState({});
  const [decisionErrors, setDecisionErrors] = useState({});
  const [notice, setNotice] = useState('');
  const [selectedEventId, setSelectedEventId] = useState(null);
  const [view, setView] = useState('upcoming');
  const [now, setNow] = useState(Date.now);
  const active = useRef(false);
  const inFlight = useRef(new Set());
  useEffect(() => {
    active.current = true;
    const update = () => setNow(Date.now());
    const timer = window.setInterval(update, 15000);
    window.addEventListener('focus', update);
    document.addEventListener('visibilitychange', update);
    return () => { active.current = false; window.clearInterval(timer); window.removeEventListener('focus', update); document.removeEventListener('visibilitychange', update); };
  }, []);
  async function decide(eventId, bookingId, status) {
    if (inFlight.current.has(bookingId)) return;
    inFlight.current.add(bookingId);
    setPending((current) => ({ ...current, [bookingId]: true }));
    setDecisionErrors((current) => ({ ...current, [bookingId]: '' }));
    setNotice('');
    try {
      const updated = await decideBooking(bookingId, status);
      if (!active.current) return;
      updateItems((items) => items.map((event) => {
        if (event.id !== eventId) return event;
        const previous = event.bookings.find((booking) => booking.id === bookingId);
        const count = event.participants_count + (updated.status === 'confirmed' ? 1 : 0) - (previous?.status === 'confirmed' ? 1 : 0);
        return { ...event, participants_count: count, is_full: event.capacity > 0 && count >= event.capacity, bookings: event.bookings.map((booking) => booking.id === bookingId ? { ...booking, ...updated } : booking) };
      }));
      setNotice(updated.status === 'confirmed' ? 'Заявка принята.' : 'Заявка отклонена.');
      await reload();
    } catch (err) {
      if (!active.current) return;
      if (err.status === 401) await refreshSession();
      if (!active.current) return;
      const text = typeof err.data === 'string' ? err.data.trim() : err.message.trim();
      setDecisionErrors((current) => ({ ...current, [bookingId]: errors[text] || 'Не удалось сохранить решение. Попробуйте ещё раз.' }));
      if (err.status === 409 || err.status === 404) await reload();
    } finally {
      inFlight.current.delete(bookingId);
      if (active.current) setPending((current) => ({ ...current, [bookingId]: false }));
    }
  }
  const groups = groupOrganizerEvents(events, now);
  const visibleEvents = groups[view];
  const requests = groups.upcoming.reduce((count, event) => count + (event.status === 'active' && Date.parse(event.starts_at) > now ? event.bookings.filter((booking) => booking.status === 'pending').length : 0), 0);
  return (
    <section className="profile-panel organizer-events">
      <header className="profile-panel-head"><div><h2>{roomId ? 'Заявки на мероприятия комнаты' : 'Я организатор'}</h2><p>{roomId ? 'Выберите мероприятие этой комнаты, чтобы посмотреть заявки участников.' : 'Созданные мероприятия и заявки участников.'}{requests > 0 && ` Ожидают решения: ${requests}.`}</p></div><button className="icon-button" type="button" aria-label="Обновить мероприятия организатора" disabled={loading} onClick={reload}><RefreshCw size={17} className={loading ? 'spin' : ''} /></button></header>
      <div className="profile-tabs organizer-event-tabs" role="group" aria-label="Период мероприятий организатора">
        {[{ key: 'upcoming', label: 'Предстоящие и текущие', icon: CalendarDays }, { key: 'past', label: 'Прошедшие', icon: History }].map(({ key, label, icon: Icon }) => <button key={key} type="button" className={`profile-tab ${view === key ? 'is-active' : ''}`} aria-pressed={view === key} onClick={() => { setView(key); setSelectedEventId(null); }}><Icon size={16} /><span>{label}</span><span className="profile-tab-count">{groups[key].length}</span></button>)}
      </div>
      {notice && <div className="event-created-note organizer-notice" role="status">{notice}</div>}
      {error && <div className="error-message profile-events-error" role="alert">{error}</div>}
      {loading && !events.length ? <div className="profile-panel-empty" role="status"><div className="loader" /><span>Загружаем мероприятия…</span></div> : !visibleEvents.length && (!error || events.length > 0) ? <div className="profile-panel-empty"><CalendarDays size={24} /><strong>{view === 'past' ? 'Прошедших мероприятий пока нет' : 'Предстоящих и текущих мероприятий пока нет'}</strong><span>{view === 'past' ? 'Здесь появятся созданные вами мероприятия после их окончания.' : roomId ? 'Нажмите «Создать мероприятие», чтобы добавить встречу.' : 'Создайте встречу внутри своей комнаты.'}</span></div> : visibleEvents.map((event) => (
        <article key={event.id} className={`organizer-event ${selectedEventId === event.id ? 'is-expanded' : ''}`}>
          <button type="button" className="organizer-event-toggle" aria-expanded={selectedEventId === event.id} aria-controls={`event-requests-${event.id}`} onClick={() => setSelectedEventId((current) => current === event.id ? null : event.id)}>
            <span className="organizer-event-summary"><strong className="organizer-event-title">{event.title}</strong><span className="organizer-event-meta">{format(event.starts_at)} — {format(event.ends_at)}<br />{event.room.name} · {event.capacity > 0 ? `${event.participants_count} из ${event.capacity} мест занято` : 'Без ограничения мест'}</span>
              <span className="organizer-request-count">Заявок: {event.bookings.length}{event.bookings.some((booking) => booking.status === 'pending') && <span> · {view === 'past' ? 'Без решения' : 'Ожидают'}: {event.bookings.filter((booking) => booking.status === 'pending').length}</span>}</span>
            </span>
            <ChevronDown size={20} className="organizer-event-chevron" />
          </button>
          <div id={`event-requests-${event.id}`} hidden={selectedEventId !== event.id}>
          {!event.bookings.length ? <p className="event-booking-hint">Заявок пока нет.</p> : <ul className="organizer-requests">{event.bookings.map((booking) => {
            const canDecide = booking.status === 'pending' && event.status === 'active' && new Date(event.starts_at).getTime() > now;
            return <li key={booking.id} className="organizer-request">
              <div className="organizer-request-row"><div className="organizer-applicant"><Avatar name={booking.username} src={booking.avatar} size="sm" /><strong>{booking.username}</strong></div>
                <div className="organizer-decision"><span className={`profile-booking-status is-${booking.status}`}>{labels[booking.status] || booking.status}</span>
                  {canDecide && <><button type="button" className="secondary-button" disabled={loading || pending[booking.id] || event.is_full} aria-label={`Принять заявку: ${booking.username}`} onClick={() => decide(event.id, booking.id, 'confirmed')}><Check size={16} />Принять</button><button type="button" className="secondary-button danger" disabled={loading || pending[booking.id]} aria-label={`Отклонить заявку: ${booking.username}`} onClick={() => decide(event.id, booking.id, 'rejected')}><X size={16} />Отклонить</button></>}
                </div></div>
              {booking.status === 'pending' && !canDecide && <p className="event-booking-hint">Рассмотрение заявок закрыто.</p>}
              {canDecide && event.is_full && <p className="event-booking-hint">Все места заняты. Принять ещё одну заявку нельзя.</p>}
              {decisionErrors[booking.id] && <div className="error-message" role="alert">{decisionErrors[booking.id]}</div>}
            </li>;
          })}</ul>}
          </div>
        </article>
      ))}
    </section>
  );
}
