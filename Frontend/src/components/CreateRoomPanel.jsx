import { Plus, X } from 'lucide-react';
import { useState } from 'react';

export default function CreateRoomPanel({ open, onClose, onSubmit, loading }) {
  const [name, setName] = useState('');

  if (!open) return null;

  async function submit(event) {
    event.preventDefault();
    if (!name.trim() || loading) return;
    const created = await onSubmit(name);
    if (created) setName('');
  }

  return (
    <div className="modal-backdrop" onMouseDown={onClose}>
      <section className="create-panel" onMouseDown={(event) => event.stopPropagation()}>
        <button className="icon-button panel-close" onClick={onClose} type="button" aria-label="Закрыть">
          <X size={20} />
        </button>
        <span className="eyebrow">Новая комната</span>
        <h2>Создай пространство для разговора</h2>
        <p>Название можно изменить позже. После создания ты сразу подключишься к комнате.</p>
        <form onSubmit={submit}>
          <label htmlFor="room-name">Название</label>
          <input
            id="room-name"
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Например, Ночной созвон"
            maxLength={64}
            autoFocus
          />
          <button className="primary-button full" disabled={!name.trim() || loading} type="submit">
            <Plus size={18} />
            {loading ? 'Создаём…' : 'Создать и войти'}
          </button>
        </form>
      </section>
    </div>
  );
}
