import { CalendarCheck, CalendarClock, CalendarDays, CalendarX, History, RefreshCw } from 'lucide-react';
import { useEffect, useState } from 'react';
import { getMyBookings } from '../api/http';
import useProfileEventList from '../hooks/useProfileEventList';
import { groupProfileBookings } from '../utils/bookings';

const TABS = [
  {
    key: 'pending',
    label: 'Ожидают',
    icon: CalendarClock,
    empty: 'Нет записей, которые ждут подтверждения.',
  },
  {
    key: 'confirmed',
    label: 'Приняты',
    icon: CalendarCheck,
    empty: 'Подтверждённых записей пока нет.',
  },
  {
    key: 'rejected',
    label: 'Отказы',
    icon: CalendarX,
    empty: 'Отказов нет — и это хорошо.',
  },
  {
    key: 'past',
    label: 'Прошедшие',
    icon: History,
    empty: 'Прошедших мероприятий пока нет.',
  },
];

function formatDateTime(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' });
}

export default function ProfileEvents({ refreshSession }) {
  const { items: bookings, loading, error, reload } = useProfileEventList(getMyBookings, refreshSession);
  const [active, setActive] = useState('pending');
  const [now, setNow] = useState(Date.now);
  useEffect(() => {
    const update = () => setNow(Date.now());
    const timer = window.setInterval(update, 15000);
    window.addEventListener('focus', update);
    document.addEventListener('visibilitychange', update);
    return () => {
      window.clearInterval(timer);
      window.removeEventListener('focus', update);
      document.removeEventListener('visibilitychange', update);
    };
  }, []);
  const tab = TABS.find((item) => item.key === active);
  const groups = groupProfileBookings(bookings, now);
  const items = groups[active];
  const countFor = (key) => groups[key].length;

  return (
    <section className="profile-panel">
      <header className="profile-panel-head">
        <div>
          <h2>Мои мероприятия</h2>
          <p>Записи, на которые вы откликнулись, и их статус.</p>
        </div>
        <button type="button" className="icon-button" aria-label="Обновить мои заявки" disabled={loading} onClick={reload}><RefreshCw size={17} className={loading ? 'spin' : ''} /></button>
      </header>

      <div className="profile-tabs profile-event-tabs" role="tablist" aria-label="Статус записей">
        {TABS.map(({ key, label, icon: Icon }) => (
          <button
            key={key}
            className={`profile-tab ${active === key ? 'is-active' : ''}`}
            onClick={() => setActive(key)}
            type="button"
            role="tab"
            aria-selected={active === key}
          >
            <Icon size={16} />
            <span>{label}</span>
            <span className="profile-tab-count">{countFor(key)}</span>
          </button>
        ))}
      </div>

      {error && <div className="error-message profile-events-error" role="alert">{error}</div>}
      {loading && !bookings.length ? <div className="profile-panel-empty" role="status"><div className="loader" /><span>Загружаем заявки…</span></div> : error && !bookings.length ? null : items.length === 0 ? (
        <div className="profile-panel-empty">
          <span className="empty-icon"><CalendarDays size={24} /></span>
          <strong>{tab.empty}</strong>
          <span>{active === 'past' ? 'Здесь появятся завершившиеся мероприятия с подтверждённой записью.' : 'Когда вы запишетесь на мероприятие, оно появится здесь.'}</span>
        </div>
      ) : (
        <ul className="profile-booking-list">
          {items.map((booking) => (
            <li key={booking.id} className="profile-booking">
              <div>
                <strong>{booking.event.title}</strong>
                <span>{formatDateTime(booking.event.starts_at)} — {formatDateTime(booking.event.ends_at)}</span>
                <span>{booking.event.room.name}</span>
              </div>
              <span className={`profile-booking-status is-${active === 'past' ? 'past' : booking.status}`}>{active === 'past' ? 'Завершилось' : tab.label}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
