import { useCallback, useEffect, useRef, useState } from 'react';
import { getMyBookings, getMyEvents, joinEvent } from '../api/http';

const joinErrors = {
  'already booked': 'Вы уже отправили заявку на это мероприятие.',
  'event already started': 'Мероприятие уже началось. Запись закрыта.',
  'no seats left': 'Свободных мест больше нет.',
  'event not found': 'Мероприятие больше недоступно.',
  'internal error': 'Не удалось отправить заявку. Попробуйте ещё раз.',
};

export default function useEventBookings(refreshSession) {
  const [bookings, setBookings] = useState({});
  const [organized, setOrganized] = useState({});
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [pending, setPending] = useState({});
  const [errors, setErrors] = useState({});
  const inFlight = useRef(new Set());
  const version = useRef(0);
  const active = useRef(false);

  const reload = useCallback(async () => {
    const requestVersion = ++version.current;
    setLoading(true);
    setLoadError('');
    try {
      const [items, ownedEvents] = await Promise.all([getMyBookings(), getMyEvents()]);
      if (!active.current || requestVersion !== version.current) return;
      setBookings(Object.fromEntries(items.map((item) => [item.event_id, item])));
      setOrganized(Object.fromEntries(ownedEvents.map((event) => [event.id, true])));
    } catch (err) {
      if (!active.current || requestVersion !== version.current) return;
      if (err.status === 401) await refreshSession();
      if (active.current && requestVersion === version.current) setLoadError('Не удалось проверить ваши заявки и мероприятия. Повторите загрузку перед записью.');
    } finally {
      if (active.current && requestVersion === version.current) setLoading(false);
    }
  }, [refreshSession]);

  useEffect(() => {
    active.current = true;
    reload();
    const visible = () => { if (document.visibilityState === 'visible') reload(); };
    window.addEventListener('focus', reload);
    document.addEventListener('visibilitychange', visible);
    return () => { active.current = false; version.current++; window.removeEventListener('focus', reload); document.removeEventListener('visibilitychange', visible); };
  }, [reload]);

  async function join(event) {
    if (loading || loadError || organized[event.id] || bookings[event.id] || inFlight.current.has(event.id)) return;
    if (new Date(event.starts_at).getTime() <= Date.now()) {
      setErrors((current) => ({ ...current, [event.id]: joinErrors['event already started'] }));
      return;
    }
    inFlight.current.add(event.id);
    setPending((current) => ({ ...current, [event.id]: true }));
    setErrors((current) => ({ ...current, [event.id]: '' }));
    try {
      const booking = await joinEvent(event.id);
      if (!active.current) return;
      version.current++;
      setLoading(false);
      setBookings((current) => ({ ...current, [event.id]: booking }));
    } catch (err) {
      if (!active.current) return;
      if (err.status === 401) await refreshSession();
      if (!active.current) return;
      const message = typeof err.data === 'string' ? err.data.trim() : err.message.trim();
      setErrors((current) => ({ ...current, [event.id]: joinErrors[message] || 'Не удалось отправить заявку. Проверьте соединение и попробуйте ещё раз.' }));
      if (err.status === 409) await reload();
    } finally {
      inFlight.current.delete(event.id);
      if (active.current) setPending((current) => ({ ...current, [event.id]: false }));
    }
  }
  return { bookings, organized, loading, loadError, pending, errors, reload, join };
}
