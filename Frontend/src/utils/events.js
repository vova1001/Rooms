export function formatEventTimeRange(startsAt, endsAt, timeZone) {
  const start = startsAt ? new Date(startsAt) : new Date(NaN);
  const end = endsAt ? new Date(endsAt) : new Date(NaN);
  if (!Number.isFinite(start.getTime()) || !Number.isFinite(end.getTime())) return 'Дата уточняется';
  const dayFormatter = new Intl.DateTimeFormat('en-CA', { timeZone, year: 'numeric', month: '2-digit', day: '2-digit' });
  const timeFormatter = new Intl.DateTimeFormat('ru-RU', { timeZone, hour: '2-digit', minute: '2-digit', hourCycle: 'h23' });
  if (dayFormatter.format(start) === dayFormatter.format(end)) {
    return `${timeFormatter.format(start)}–${timeFormatter.format(end)}`;
  }
  const yearFormatter = new Intl.DateTimeFormat('en', { timeZone, year: 'numeric' });
  const showYear = yearFormatter.format(start) !== yearFormatter.format(end);
  const dateFormatter = new Intl.DateTimeFormat('ru-RU', { timeZone, day: 'numeric', month: 'long', ...(showYear ? { year: 'numeric' } : {}) });
  return `${dateFormatter.format(start)}, ${timeFormatter.format(start)} — ${dateFormatter.format(end)}, ${timeFormatter.format(end)}`;
}

// Date сам исправляет 31 февраля; для формы такое исправление недопустимо.
export function parseEventLocalDateTime(value) {
  const match = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value);
  if (!match) return null;
  const [, year, month, day, hour, minute] = match.map(Number);
  const date = new Date(value);
  if (!Number.isFinite(date.getTime()) || date.getFullYear() !== year || date.getMonth() + 1 !== month || date.getDate() !== day || date.getHours() !== hour || date.getMinutes() !== minute) return null;
  return date;
}

export function parseTypedEventDateTime(text) {
  const match = /^(\d{2})\.(\d{2})\.(\d{4}) (\d{2}):(\d{2})$/.exec(text);
  if (!match) return null;
  const [, day, month, year, hour, minute] = match;
  const value = `${year}-${month}-${day}T${hour}:${minute}`;
  return parseEventLocalDateTime(value) ? value : null;
}

export function formatEventDateTimeInput(value) {
  if (!parseEventLocalDateTime(value)) return value;
  const [date, time] = value.split('T');
  const [year, month, day] = date.split('-');
  return `${day}.${month}.${year} ${time}`;
}

export function maskEventDateTime(text, caret, previous = '', inputType = '') {
  let digits = text.replace(/\D/g, '');
  let beforeCaret = text.slice(0, caret).replace(/\D/g, '').length;
  // Backspace/Delete на разделителе должны удалять цифру, а не возвращать разделитель обратно.
  if (digits === previous.replace(/\D/g, '') && text.length < previous.length) {
    const index = inputType === 'deleteContentBackward' ? beforeCaret - 1 : beforeCaret;
    if (index >= 0 && index < digits.length) {
      digits = digits.slice(0, index) + digits.slice(index + 1);
      if (inputType === 'deleteContentBackward') beforeCaret = index;
    }
  }
  digits = digits.slice(0, 12);
  const separators = { 2: '.', 4: '.', 8: ' ', 10: ':' };
  let formatted = '';
  for (let index = 0; index < digits.length; index++) {
    formatted += digits[index];
    if (separators[index + 1]) formatted += separators[index + 1];
  }
  let position = 0;
  let seen = 0;
  while (position < formatted.length && seen < beforeCaret) {
    if (/\d/.test(formatted[position])) seen++;
    position++;
  }
  while (position < formatted.length && /\D/.test(formatted[position])) position++;
  return { text: formatted, caret: position };
}

export function buildCreateEventPayload(fields, now = Date.now()) {
  const title = fields.title.trim();
  const description = fields.description.trim();
  if (!title || [...title].length > 200) throw new Error('Название — от 1 до 200 символов.');
  if ([...description].length > 2000) throw new Error('Описание — не больше 2000 символов.');
  const start = parseEventLocalDateTime(fields.start);
  const end = parseEventLocalDateTime(fields.end);
  if (!start || !end) {
    throw new Error('Укажите корректные начало и окончание: ДД.ММ.ГГГГ ЧЧ:ММ.');
  }
  if (start.getTime() <= now) throw new Error('Начало мероприятия должно быть в будущем.');
  if (end <= start) throw new Error('Окончание должно быть позже начала.');
  if (end - start > 12 * 60 * 60 * 1000) throw new Error('Мероприятие может длиться не больше 12 часов.');

  let priceMinor = 0;
  if (!fields.isFree) {
    const price = fields.price.trim().replace(',', '.');
    if (!/^\d+(\.\d{1,2})?$/.test(price)) throw new Error('Введите цену в рублях, не больше двух знаков после запятой.');
    const [rubles, kopecks = ''] = price.split('.');
    priceMinor = Number(rubles) * 100 + Number(kopecks.padEnd(2, '0'));
    if (!Number.isSafeInteger(priceMinor) || priceMinor <= 0 || priceMinor > 100000000) {
      throw new Error('Цена — от 0,01 до 1 000 000 рублей.');
    }
  }
  const capacity = fields.capacity.trim();
  if (!/^\d+$/.test(capacity) || Number(capacity) > 1000) throw new Error('Число мест — целое от 0 до 1000. Ноль означает без ограничений.');
  return {
    title, description, starts_at: start.toISOString(), ends_at: end.toISOString(),
    is_free: fields.isFree, price_minor: priceMinor, capacity: Number(capacity),
  };
}
