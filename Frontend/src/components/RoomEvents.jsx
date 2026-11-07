import EventsPanel from './EventsPanel';
import { useEffect, useRef, useState } from 'react';
import { Plus } from 'lucide-react';
import { createEvent, getRooms } from '../api/http';
import CreateEventPanel from './CreateEventPanel';
import OrganizerEvents from './OrganizerEvents';
import { normalizeRoom } from '../utils/rooms';

export default function RoomEvents({ room, user, refreshSession }) {
  const [ownerId, setOwnerId] = useState(room.ownerId);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [created, setCreated] = useState(null);
  const [view, setView] = useState('calendar');
  const isOwner = Boolean(ownerId && ownerId === user?.id);
  const inFlight = useRef(false);
  useEffect(() => {
    let active = true;
    setOwnerId(room.ownerId);
    // При прямом входе по ссылке sessionStorage может не содержать владельца.
    if (!room.ownerId) getRooms().then((rooms) => {
      if (active) setOwnerId(rooms.map(normalizeRoom).find((item) => item.id === room.id)?.ownerId);
    }).catch(async (err) => { if (active && err.status === 401) await refreshSession(); });
    return () => { active = false; };
  }, [room.id, room.ownerId, refreshSession]);
  async function submit(fields) {
    if (inFlight.current) return;
    inFlight.current = true;
    setSaving(true);
    setError('');
    try {
      const event = await createEvent(room.id, fields);
      setCreated(event);
      setOpen(false);
      setView('calendar');
    } catch (err) {
      if (err.status === 401) await refreshSession();
      setError(err.status === 404 ? 'Комната не найдена или у вас нет прав на создание мероприятия.' : err.message.replace(/^validation error:\s*/i, ''));
    } finally { inFlight.current = false; setSaving(false); }
  }
  return (
    <section className="room-events-section">
      <header className="room-events-heading"><div><span className="eyebrow">Встречи в этом пространстве</span><h2>Мероприятия комнаты</h2><p>Расписание «{room.name}» — выберите день и найдите свою следующую встречу.</p></div>
        {isOwner && <button type="button" className="primary-button" onClick={() => { setError(''); setOpen(true); }}><Plus size={18} />Создать мероприятие</button>}
      </header>
      {created && <div className="event-created-note" role="status">Мероприятие «{created.title}» создано.</div>}
      {isOwner && <nav className="home-tabs room-event-tabs" aria-label="Управление мероприятиями комнаты">
        <button type="button" className={view === 'calendar' ? 'is-active' : ''} aria-pressed={view === 'calendar'} onClick={() => setView('calendar')}>Календарь</button>
        <button type="button" className={view === 'requests' ? 'is-active' : ''} aria-pressed={view === 'requests'} onClick={() => setView('requests')}>Заявки участников</button>
      </nav>}
      {isOwner && view === 'requests'
        ? <OrganizerEvents key={room.id} roomId={room.id} refreshSession={refreshSession} />
        : <EventsPanel key={created?.id || room.id} roomId={room.id} refreshSession={refreshSession} initialDate={created?.starts_at} />}
      {open && <CreateEventPanel room={room} saving={saving} error={error} onClose={() => { if (!inFlight.current) setOpen(false); }} onSubmit={submit} />}
    </section>
  );
}
