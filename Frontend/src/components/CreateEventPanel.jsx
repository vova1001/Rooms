import { useEffect, useRef, useState } from 'react';
import { CalendarPlus, X } from 'lucide-react';
import { buildCreateEventPayload } from '../utils/events';
import EventDateTimeField from './EventDateTimeField';

export default function CreateEventPanel({ room, onClose, onSubmit, saving, error }) {
  const dialog = useRef(null);
  const [fields, setFields] = useState({ title: '', description: '', start: '', end: '', isFree: true, price: '', capacity: '0' });
  const [validation, setValidation] = useState('');
  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  useEffect(() => {
    const element = dialog.current;
    element.showModal();
    return () => element.close();
  }, []);
  const change = (key, value) => { setFields((current) => ({ ...current, [key]: value })); setValidation(''); };
  async function submit(event) {
    event.preventDefault();
    if (saving) return;
    let payload;
    try { payload = buildCreateEventPayload(fields); }
    catch (err) { setValidation(err.message); return; }
    await onSubmit(payload);
  }
  return (
    <dialog ref={dialog} className="create-panel create-event-dialog" aria-labelledby="create-event-title" onCancel={(event) => { event.preventDefault(); if (!saving) onClose(); }}>
      <button className="icon-button panel-close" type="button" aria-label="Закрыть форму" disabled={saving} onClick={onClose}><X size={20} /></button>
      <span className="eyebrow">Новая встреча</span><h2 id="create-event-title">Создать мероприятие</h2>
      <p>Комната «{room.name}». Время указано в часовом поясе {timeZone}.</p>
      <form onSubmit={submit} noValidate>
        <label htmlFor="event-title">Название</label>
        <input id="event-title" value={fields.title} onChange={(event) => change('title', event.target.value)} placeholder="Например, Разбор React" disabled={saving} />
        <label htmlFor="event-description">Описание</label>
        <textarea id="event-description" rows={3} value={fields.description} onChange={(event) => change('description', event.target.value)} placeholder="Что будем обсуждать?" disabled={saving} />
        <div className="event-form-columns">
          <EventDateTimeField id="event-start" label="Начало" value={fields.start} onChange={(value) => change('start', value)} disabled={saving} />
          <EventDateTimeField id="event-end" label="Окончание" value={fields.end} onChange={(value) => change('end', value)} disabled={saving} />
        </div>
        <small className="event-form-hint">До 12 часов, в том числе с переходом на следующий день.</small>
        <fieldset className="event-price-choice" disabled={saving}><legend>Участие</legend>
          <label><input type="radio" name="event-payment" checked={fields.isFree} onChange={() => change('isFree', true)} /><span>Бесплатное</span></label>
          <label><input type="radio" name="event-payment" checked={!fields.isFree} onChange={() => change('isFree', false)} /><span>Платное</span></label>
        </fieldset>
        {!fields.isFree && <><label htmlFor="event-price">Стоимость, ₽</label><input id="event-price" inputMode="decimal" value={fields.price} onChange={(event) => change('price', event.target.value)} placeholder="Например, 1500" disabled={saving} /><small className="event-form-hint">Стоимость для участников. Оплата в приложении пока не принимается.</small></>}
        <label htmlFor="event-capacity">Количество мест</label><input id="event-capacity" type="number" min="0" max="1000" step="1" value={fields.capacity} onChange={(event) => change('capacity', event.target.value.replace(/^0+(?=\d)/, ''))} disabled={saving} />
        <small className="event-form-hint">0 — без ограничений.</small>
        {(validation || error) && <div className="error-message" role="alert">{validation || error}</div>}
        <button className="primary-button full" type="submit" disabled={saving}><CalendarPlus size={18} />{saving ? 'Создаём…' : 'Создать мероприятие'}</button>
      </form>
    </dialog>
  );
}
