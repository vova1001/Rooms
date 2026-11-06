import { useEffect, useState } from 'react';
import { CalendarPlus, Check, Clock, X } from 'lucide-react';

const statuses = {
  pending: { label: 'Заявка на рассмотрении', Icon: Clock },
  confirmed: { label: 'Вы записаны на мероприятие', Icon: Check },
  rejected: { label: 'Заявка отклонена', Icon: X },
};

export default function EventBookingAction({ event, booking, isOrganizer, loading, sending, error, onJoin }) {
  const [now, setNow] = useState(Date.now);
  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 15000);
    return () => window.clearInterval(timer);
  }, []);
  const started = new Date(event.starts_at).getTime() <= now;
  const full = event.is_full || (event.capacity > 0 && event.participants_count >= event.capacity);
  const status = booking && statuses[booking.status];
  return (
    <div className="event-booking-area">
      <div className="event-booking-row">
        <span className="event-seat-count">{event.capacity > 0 ? `${event.participants_count} из ${event.capacity} мест занято` : 'Без ограничения мест'}</span>
        {isOrganizer ? <span className="event-booking-status is-confirmed"><Check size={16} />Вы организатор</span>
          : status ? <span className={`event-booking-status is-${booking.status}`} role="status"><status.Icon size={16} />{status.label}</span>
          : booking ? <span className="event-booking-status" role="status">Заявка уже отправлена</span>
          : <button className="primary-button" type="button" disabled={loading || sending || started || full} onClick={() => onJoin(event)}><CalendarPlus size={17} />{sending ? 'Отправляем…' : started ? 'Запись закрыта' : full ? 'Мест нет' : loading ? 'Проверяем заявки…' : 'Подать заявку'}</button>}
      </div>
      {!isOrganizer && !booking && !started && !full && !loading && <p className="event-booking-hint">Заявку рассмотрит организатор.{!event.is_free && ' Подача заявки не является оплатой.'}</p>}
      {!isOrganizer && error && <div className="error-message" role="alert">{error}</div>}
    </div>
  );
}
