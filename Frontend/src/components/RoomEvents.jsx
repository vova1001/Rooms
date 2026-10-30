import EventsPanel from './EventsPanel';

export default function RoomEvents({ room, refreshSession }) {
  return (
    <section className="room-events-section">
      <header className="room-events-heading"><span className="eyebrow">Встречи в этом пространстве</span><h2>Мероприятия комнаты</h2><p>Расписание «{room.name}» — выберите день и найдите свою следующую встречу.</p></header>
      <EventsPanel roomId={room.id} refreshSession={refreshSession} />
    </section>
  );
}
