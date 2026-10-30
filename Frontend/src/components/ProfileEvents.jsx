import { CalendarCheck, CalendarClock, CalendarDays, CalendarX } from 'lucide-react';
import { useState } from 'react';

// Статусы записей на мероприятия. Ключи совпадают с bookings.status на сервере.
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
];

function formatDateTime(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('ru-RU', { day: 'numeric', month: 'long', hour: '2-digit', minute: '2-digit' });
}

// bookings: [{ id, title, start_at, status }] — когда появится ручка на бэкенде,
// сюда достаточно передать её ответ.
export default function ProfileEvents({ bookings = [] }) {
  const [active, setActive] = useState('pending');
  const tab = TABS.find((item) => item.key === active);
  const items = bookings.filter((booking) => booking.status === active);
  const countFor = (key) => bookings.filter((booking) => booking.status === key).length;

  return (
    <section className="profile-panel">
      <header className="profile-panel-head">
        <div>
          <h2>Мои мероприятия</h2>
          <p>Записи, на которые вы откликнулись, и их статус.</p>
        </div>
      </header>

      <div className="profile-tabs" role="tablist" aria-label="Статус записей">
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

      {items.length === 0 ? (
        <div className="profile-panel-empty">
          <span className="empty-icon"><CalendarDays size={24} /></span>
          <strong>{tab.empty}</strong>
          <span>Когда вы запишетесь на мероприятие, оно появится здесь.</span>
        </div>
      ) : (
        <ul className="profile-booking-list">
          {items.map((booking) => (
            <li key={booking.id} className="profile-booking">
              <div>
                <strong>{booking.title}</strong>
                <span>{formatDateTime(booking.start_at)}</span>
              </div>
              <span className={`profile-booking-status is-${booking.status}`}>{tab.label}</span>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
