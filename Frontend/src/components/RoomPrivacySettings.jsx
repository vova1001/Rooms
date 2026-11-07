import { useEffect, useRef, useState } from 'react';
import { Copy, LockKeyhole, Settings2, X } from 'lucide-react';
import { updateRoomPrivacy } from '../api/http';
import { generateRoomCode } from '../utils/rooms';
import { readRoomCode, storeRoomCode } from '../utils/roomCodes';

function PrivacyDialog({ room, refreshSession, onSaved, onClose }) {
  const dialog = useRef(null);
  const inFlight = useRef(false);
  const active = useRef(false);
  const codeInput = useRef(null);
  const [isPrivate, setIsPrivate] = useState(room.isPrivate);
  const [password, setPassword] = useState(() => readRoomCode(room));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);
  const [storageNotice, setStorageNotice] = useState('');
  const [copyNotice, setCopyNotice] = useState('');
  function changeCode(value) {
    setPassword(value);
    setSaved(false);
    setError('');
    setCopyNotice('');
  }
  async function copyCode() {
    const value = password;
    try {
      await navigator.clipboard.writeText(value);
      if (active.current && codeInput.current?.value === value) setCopyNotice('Код скопирован.');
    } catch {
      if (!active.current || codeInput.current?.value !== value) return;
      codeInput.current?.focus();
      codeInput.current?.select();
      setCopyNotice('Не удалось скопировать автоматически. Код выделен — скопируйте его вручную.');
    }
  }
  useEffect(() => {
    active.current = true;
    const element = dialog.current;
    element.showModal();
    return () => { active.current = false; element.close(); };
  }, []);
  async function submit(event) {
    event.preventDefault();
    if (inFlight.current || isPrivate === room.isPrivate) return;
    if (isPrivate && !password) {
      setError('Код этой комнаты недоступен в этом браузере.');
      return;
    }
    inFlight.current = true;
    setSaving(true);
    setError('');
    try {
      await updateRoomPrivacy(room.id, isPrivate, password);
      const stored = storeRoomCode(room, isPrivate, password);
      if (active.current) {
        setSaved(true);
        setStorageNotice(stored ? '' : 'Настройки сохранены на сервере, но браузер не смог запомнить код. Скопируйте его перед закрытием формы.');
        onSaved(isPrivate);
      }
    } catch (err) {
      if (!active.current) return;
      if (err.status === 401) await refreshSession();
      if (active.current) setError(err.status === 403 || err.status === 404 ? 'Комната не найдена или у вас нет прав на её изменение.' : 'Не удалось сохранить настройки доступа. Попробуйте ещё раз.');
    } finally {
      inFlight.current = false;
      if (active.current) setSaving(false);
    }
  }
  return <dialog ref={dialog} className="create-panel create-event-dialog room-privacy-dialog" aria-labelledby="room-privacy-title" onCancel={(event) => { event.preventDefault(); if (!inFlight.current) onClose(); }}>
    <button className="icon-button panel-close" type="button" aria-label="Закрыть настройки" disabled={saving} onClick={onClose}><X size={20} /></button>
    <span className="eyebrow">Настройки комнаты</span><h2 id="room-privacy-title">Доступ в «{room.name}»</h2>
    <form onSubmit={submit}>
      <fieldset className="event-price-choice" disabled={saving}><legend>Кто может войти</legend>
        <label><input type="radio" name="room-privacy" checked={!isPrivate} onChange={() => { setIsPrivate(false); setSaved(false); setError(''); }} /><span>Все</span></label>
        <label><input type="radio" name="room-privacy" checked={isPrivate} onChange={() => { setIsPrivate(true); changeCode(password || readRoomCode(room) || (!room.isPrivate ? generateRoomCode() : '')); }} /><span>По коду</span></label>
      </fieldset>
      {isPrivate ? <><label htmlFor="room-new-code">Код комнаты</label>
        <input ref={codeInput} id="room-new-code" type="text" autoComplete="off" spellCheck={false} value={password} readOnly placeholder="Код недоступен в этом браузере" onClick={(event) => event.currentTarget.select()} />
        <div className="room-code-actions">
          <button className="secondary-button" type="button" disabled={saving || !password} onClick={copyCode}><Copy size={16} />Скопировать</button>
        </div>
        {copyNotice && <p className="event-form-hint" role="status">{copyNotice}</p>}
        <p className="event-form-hint">{room.isPrivate ? 'Передайте код тем, кого хотите пригласить.' : 'Код начнёт действовать после сохранения настроек.'}</p></> : <p className="event-form-hint">В комнату можно будет войти без кода.</p>}
      {error && <div className="error-message" role="alert">{error}</div>}
      {saved && <div className="event-created-note" role="status">Настройки доступа сохранены.</div>}
      {storageNotice && <p className="event-form-hint" role="status">{storageNotice}</p>}
      {isPrivate !== room.isPrivate && <button className="primary-button full" disabled={saving || (isPrivate && !password)} type="submit">{saving ? 'Сохраняем…' : 'Сохранить'}</button>}
    </form>
  </dialog>;
}

export default function RoomPrivacySettings({ room, refreshSession, onSaved }) {
  const [open, setOpen] = useState(false);
  return <>
    <button type="button" className="secondary-button" onClick={() => setOpen(true)} aria-label="Настройки доступа в комнату">{room.isPrivate ? <LockKeyhole size={17} /> : <Settings2 size={17} />}Доступ</button>
    {open && <PrivacyDialog room={room} refreshSession={refreshSession} onClose={() => setOpen(false)} onSaved={onSaved} />}
  </>;
}
