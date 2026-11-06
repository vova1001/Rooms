import { ChevronLeft, ChevronRight, CalendarDays, Clock, Lock } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getEvents } from '../api/http';
import useEventBookings from '../hooks/useEventBookings';
import EventBookingAction from './EventBookingAction';
import { formatEventTimeRange } from '../utils/events';

function dateKey(date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

export default function EventsPanel({ refreshSession, roomId, initialDate }) {
  const applications = useEventBookings(refreshSession);
  const [today] = useState(() => new Date());
  const [selected, setSelected] = useState(() => dateKey(initialDate ? new Date(initialDate) : today));
  const [month, setMonth] = useState(() => { const date = initialDate ? new Date(initialDate) : today; return new Date(date.getFullYear(), date.getMonth(), 1); });
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [retry, setRetry] = useState(0);
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError('');
    getEvents(selected, timeZone, roomId)
      .then((data) => {
        if (active) setEvents(data.items);
      })
      .catch(async (err) => {
        if (!active) return;
        if (err.status === 401) await refreshSession();
        if (active) setError(err.message || 'Не удалось загрузить мероприятия');
      })
      .finally(() => { if (active) setLoading(false); });
    return () => { active = false; };
  }, [selected, timeZone, retry, refreshSession, roomId]);

  const offset = (month.getDay() + 6) % 7;
  const count = new Date(month.getFullYear(), month.getMonth() + 1, 0).getDate();
  function moveMonth(step) {
    const next = new Date(month.getFullYear(), month.getMonth() + step, 1);
    setMonth(next);
    setSelected(dateKey(next));
  }

  return (
    <section className="events-layout" aria-label="Мероприятия">
      <aside className="events-calendar">
        <div className="calendar-header">
          <button className="icon-button" type="button" aria-label="Предыдущий месяц" onClick={() => moveMonth(-1)}><ChevronLeft size={18} /></button>
          <strong aria-live="polite">{month.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })}</strong>
          <button className="icon-button" type="button" aria-label="Следующий месяц" onClick={() => moveMonth(1)}><ChevronRight size={18} /></button>
        </div>
        <div className="calendar-grid">
          {['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'].map((day) => <span className="calendar-weekday" key={day}>{day}</span>)}
          {Array.from({ length: offset }, (_, index) => <span key={`empty-${index}`} />)}
          {Array.from({ length: count }, (_, index) => {
            const date = new Date(month.getFullYear(), month.getMonth(), index + 1);
            const key = dateKey(date);
            return <button type="button" key={key} className={`calendar-day ${key === selected ? 'is-selected' : ''} ${key === dateKey(today) ? 'is-today' : ''}`} aria-pressed={key === selected} aria-label={date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })} onClick={() => setSelected(key)}>{index + 1}</button>;
          })}
        </div>
        <button className="secondary-button full" type="button" onClick={() => { setMonth(new Date(today.getFullYear(), today.getMonth(), 1)); setSelected(dateKey(today)); }}>Сегодня</button>
        <p className="events-timezone">Часовой пояс: {timeZone}</p>
      </aside>
      <h2 className="events-date-heading">{new Date(`${selected}T12:00:00`).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })}</h2>
      <div className="events-feed" aria-busy={loading}>
        {applications.loadError && <div className="event-booking-load-error"><div className="error-message" role="alert">{applications.loadError}</div><button type="button" className="secondary-button" onClick={applications.reload}>Обновить заявки</button></div>}
        {loading ? <div className="empty-state" role="status"><div className="loader" /><p>Загружаем мероприятия…</p></div> : error ? <div className="events-error"><div className="error-banner" role="alert">{error}</div><button type="button" className="secondary-button" onClick={() => setRetry((value) => value + 1)}>Попробовать снова</button></div> : events.length ? events.map((event) => (
          <article className="event-card" key={event.id}>
            <div className="event-card-meta"><span className="event-time-range"><Clock size={15} /><span>{formatEventTimeRange(event.starts_at, event.ends_at, timeZone)}</span></span><span className="event-price">{event.is_free ? 'Бесплатно' : new Intl.NumberFormat('ru-RU', { style: 'currency', currency: event.currency }).format(event.price_minor / 100)}</span></div>
            <h3>{event.title}</h3><p>{event.description}</p>
            <div className="event-room">{event.room.visibility === 'closed' && <Lock size={14} />}{event.room.name}{event.room.visibility === 'closed' && ' · вход по коду'}</div>
            <EventBookingAction event={event} booking={applications.bookings[event.id]} isOrganizer={Boolean(applications.organized[event.id])} loading={applications.loading || Boolean(applications.loadError)} sending={Boolean(applications.pending[event.id])} error={applications.errors[event.id]} onJoin={applications.join} />
          </article>
        )) : <div className="empty-state"><CalendarDays size={30} /><h3>{roomId ? 'В комнате пока нет встреч на этот день' : 'На этот день планов нет'}</h3><p>Выберите другую дату в календаре.</p></div>}
      </div>
    </section>
  );
}
