import EventsPanel from './EventsPanel';
import { useEffect, useRef, useState } from 'react';
import { Plus } from 'lucide-react';
import { createEvent, getRooms } from '../api/http';
import CreateEventPanel from './CreateEventPanel';

export default function RoomEvents({ room, user, refreshSession }) {
  const [ownerId, setOwnerId] = useState(room.ownerId);
  const [open, setOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [created, setCreated] = useState(null);
  const inFlight = useRef(false);
  useEffect(() => {
    let active = true;
    setOwnerId(room.ownerId);
    // При прямом входе по ссылке sessionStorage может не содержать владельца.
    if (!room.ownerId) getRooms().then((rooms) => {
      if (active) setOwnerId(rooms.find((item) => (item.id || item.ID) === room.id)?.owner_id);
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
    } catch (err) {
      if (err.status === 401) await refreshSession();
      setError(err.status === 404 ? 'Комната не найдена или у вас нет прав на создание мероприятия.' : err.message.replace(/^validation error:\s*/i, ''));
    } finally { inFlight.current = false; setSaving(false); }
  }
  return (
    <section className="room-events-section">
      <header className="room-events-heading"><div><span className="eyebrow">Встречи в этом пространстве</span><h2>Мероприятия комнаты</h2><p>Расписание «{room.name}» — выберите день и найдите свою следующую встречу.</p></div>
        {ownerId && ownerId === user?.id && <button type="button" className="primary-button" onClick={() => { setError(''); setOpen(true); }}><Plus size={18} />Создать мероприятие</button>}
      </header>
      {created && <div className="event-created-note" role="status">Мероприятие «{created.title}» создано.</div>}
      <EventsPanel key={created?.id || room.id} roomId={room.id} refreshSession={refreshSession} initialDate={created?.starts_at} />
      {open && <CreateEventPanel room={room} saving={saving} error={error} onClose={() => { if (!inFlight.current) setOpen(false); }} onSubmit={submit} />}
    </section>
  );
}
