import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { CalendarDays, ChevronLeft, ChevronRight } from 'lucide-react';
import { formatEventDateTimeInput, maskEventDateTime, parseEventLocalDateTime, parseTypedEventDateTime } from '../utils/events';

const pad = (value) => String(value).padStart(2, '0');
const dateKey = (date) => `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;

export default function EventDateTimeField({ id, label, value, onChange, disabled }) {
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState('');
  const [month, setMonth] = useState(() => new Date(new Date().getFullYear(), new Date().getMonth(), 1));
  const trigger = useRef(null);
  const input = useRef(null);
  const pendingCaret = useRef(null);
  useLayoutEffect(() => {
    if (pendingCaret.current === null) return;
    input.current?.setSelectionRange(pendingCaret.current, pendingCaret.current);
    pendingCaret.current = null;
  });
  function typeDateTime(event) {
    const field = event.target;
    const masked = maskEventDateTime(field.value, field.selectionStart ?? field.value.length, formatEventDateTimeInput(value), event.nativeEvent.inputType);
    pendingCaret.current = masked.caret;
    onChange(parseTypedEventDateTime(masked.text) || masked.text);
    setOpen(false);
  }
  const hours = useRef(null);
  const minutes = useRef(null);
  const [date = '', time = '18:00'] = draft.split('T');
  const [hour, minute] = time.split(':');
  useEffect(() => {
    if (!open) return;
    for (const list of [hours.current, minutes.current]) {
      const selected = list?.querySelector('[aria-pressed="true"]');
      if (selected) list.scrollTop = selected.offsetTop - list.clientHeight / 2 + selected.clientHeight / 2;
    }
  }, [open]);
  function show() {
    if (open) { setOpen(false); return; }
    const valid = parseEventLocalDateTime(value);
    const initial = valid || new Date();
    setDraft(valid ? value : `${dateKey(initial)}T18:00`);
    setMonth(new Date(initial.getFullYear(), initial.getMonth(), 1));
    setOpen(true);
  }
  function finish(next) {
    onChange(next);
    setOpen(false);
    trigger.current?.focus();
  }
  const offset = (month.getDay() + 6) % 7;
  const count = new Date(month.getFullYear(), month.getMonth() + 1, 0).getDate();
  return (
    <div className={`event-datetime-field ${open ? 'is-open' : ''}`}>
      <label htmlFor={id}>{label}</label>
      <div className="event-date-input-row">
        <input ref={input} id={id} type="text" inputMode="numeric" autoComplete="off" placeholder="ДД.ММ.ГГГГ ЧЧ:ММ" aria-describedby={`${id}-hint`} value={formatEventDateTimeInput(value)} disabled={disabled} onChange={typeDateTime} />
        <button ref={trigger} type="button" className="event-date-trigger" disabled={disabled} aria-label={`Открыть календарь: ${label.toLowerCase()}`} aria-expanded={open} aria-controls={`${id}-picker`} onClick={show}><CalendarDays size={18} /></button>
      </div>
      <small id={`${id}-hint`} className="event-form-hint">Введите цифры — разделители появятся автоматически, или выберите в календаре.</small>
      {open && <div id={`${id}-picker`} className="event-datetime-picker" role="group" aria-label={`${label}: дата и время`} onKeyDown={(event) => { if (event.key === 'Escape') { event.preventDefault(); event.stopPropagation(); setOpen(false); trigger.current?.focus(); } }}>
        <fieldset disabled={disabled} className="event-picker-fields">
          <legend className="event-picker-sr">{label}</legend>
          <div className="event-date-calendar">
            <div className="calendar-header">
              <button type="button" className="icon-button" aria-label="Предыдущий месяц" onClick={() => setMonth(new Date(month.getFullYear(), month.getMonth() - 1, 1))}><ChevronLeft size={17} /></button>
              <strong aria-live="polite">{month.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })}</strong>
              <button type="button" className="icon-button" aria-label="Следующий месяц" onClick={() => setMonth(new Date(month.getFullYear(), month.getMonth() + 1, 1))}><ChevronRight size={17} /></button>
            </div>
            <div className="calendar-grid">
              {['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'].map((day) => <span className="calendar-weekday" key={day}>{day}</span>)}
              {Array.from({ length: offset }, (_, index) => <span key={`space-${index}`} />)}
              {Array.from({ length: count }, (_, index) => {
                const key = `${month.getFullYear()}-${pad(month.getMonth() + 1)}-${pad(index + 1)}`;
                return <button key={key} type="button" className={`calendar-day ${date === key ? 'is-selected' : ''}`} aria-label={new Date(`${key}T12:00:00`).toLocaleDateString('ru-RU')} aria-pressed={date === key} onClick={() => setDraft(`${key}T${time}`)}>{index + 1}</button>;
              })}
            </div>
          </div>
          <div className="event-picker-time">
            <div><span id={`${id}-hours-label`}>Часы</span><div ref={hours} className="event-time-list" role="group" aria-labelledby={`${id}-hours-label`}>{Array.from({ length: 24 }, (_, index) => <button type="button" key={index} aria-label={`${pad(index)} ч`} aria-pressed={hour === pad(index)} onClick={() => setDraft(`${date}T${pad(index)}:${minute}`)}>{pad(index)}</button>)}</div></div>
            <div><span id={`${id}-minutes-label`}>Минуты</span><div ref={minutes} className="event-time-list" role="group" aria-labelledby={`${id}-minutes-label`}>{Array.from({ length: 60 }, (_, index) => <button type="button" key={index} aria-label={`${pad(index)} мин`} aria-pressed={minute === pad(index)} onClick={() => setDraft(`${date}T${hour}:${pad(index)}`)}>{pad(index)}</button>)}</div></div>
          </div>
        </fieldset>
        <div className="event-picker-actions">
          <button type="button" className="secondary-button" disabled={disabled} onClick={() => { const now = new Date(); setDraft(`${dateKey(now)}T${time}`); setMonth(new Date(now.getFullYear(), now.getMonth(), 1)); }}>Сегодня</button>
          <button type="button" className="primary-button" disabled={disabled} onClick={() => finish(draft)}>Готово</button>
        </div>
      </div>}
    </div>
  );
}
