import { Plus, RefreshCw, Search } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createRoom, getRooms } from '../api/http';
import Avatar from '../components/Avatar';
import CreateRoomPanel from '../components/CreateRoomPanel';
import Logo from '../components/Logo';
import RoomCard from '../components/RoomCard';
import { getStoredUser } from '../utils/session';
import { normalizeRoom } from '../utils/rooms';

export default function HomePage({ refreshSession }) {
  const navigate = useNavigate();
  const user = getStoredUser();
  const [rooms, setRooms] = useState([]);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [panelOpen, setPanelOpen] = useState(false);
  const [error, setError] = useState('');

  async function handleUnauthorized(err) {
    if (err?.status !== 401) return false;
    const next = await refreshSession();
    if (next.status !== 'authorized') {
      navigate(next.status === 'registration' ? '/complete-profile' : '/welcome', { replace: true });
    }
    return true;
  }

  async function loadRooms() {
    setLoading(true);
    setError('');
    try {
      const data = await getRooms();
      setRooms((Array.isArray(data) ? data : []).map(normalizeRoom));
    } catch (err) {
      if (!(await handleUnauthorized(err))) setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadRooms();
  }, []);

  const filteredRooms = useMemo(() => {
    const value = query.trim().toLowerCase();
    if (!value) return rooms;
    return rooms.filter((room) => room.name.toLowerCase().includes(value));
  }, [query, rooms]);

  function joinRoom(room) {
    sessionStorage.setItem(`room:${room.id}`, JSON.stringify(room));
    navigate(`/rooms/${room.id}`);
  }

  async function handleCreate(name) {
    setCreating(true);
    setError('');
    try {
      const created = normalizeRoom(await createRoom(name, user?.id));
      setRooms((current) => [created, ...current.filter((room) => room.id !== created.id)]);
      setPanelOpen(false);
      joinRoom(created);
      return true;
    } catch (err) {
      if (!(await handleUnauthorized(err))) setError(err.message);
      return false;
    } finally {
      setCreating(false);
    }
  }

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <Logo />
        <div className="sidebar-spacer" />
        <div className="profile-block">
          <Avatar name={user?.username} src={user?.avatar} />
          <div><strong>{user?.username || 'Пользователь'}</strong><span>В сети</span></div>
        </div>
        <div className="session-note">Вход защищён серверной сессией</div>
      </aside>

      <main className="home-main">
        <header className="home-header">
          <div>
            <span className="eyebrow">Открытые комнаты</span>
            <h1>Куда зайдём?</h1>
          </div>
          <button className="primary-button" onClick={() => setPanelOpen(true)} type="button">
            <Plus size={18} /> Создать комнату
          </button>
        </header>

        <section className="toolbar">
          <label className="search-box">
            <Search size={19} />
            <input value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Поиск комнат" />
          </label>
          <button className="icon-button" onClick={loadRooms} type="button" aria-label="Обновить">
            <RefreshCw size={19} className={loading ? 'spin' : ''} />
          </button>
        </section>

        {error && <div className="error-banner">{error}</div>}

        <section className="room-grid">
          {!loading && filteredRooms.map((room) => (
            <RoomCard key={room.id} room={room} onJoin={joinRoom} />
          ))}
        </section>

        {loading && <div className="empty-state"><div className="loader" /><h2>Ищем комнаты</h2></div>}
        {!loading && !filteredRooms.length && (
          <div className="empty-state">
            <span className="empty-icon"><Plus size={25} /></span>
            <h2>{query ? 'Ничего не найдено' : 'Пока здесь тихо'}</h2>
            <p>{query ? 'Попробуйте другое название.' : 'Создайте первую комнату и пригласите людей.'}</p>
            {!query && <button className="secondary-button" onClick={() => setPanelOpen(true)} type="button">Создать комнату</button>}
          </div>
        )}
      </main>

      <CreateRoomPanel open={panelOpen} onClose={() => setPanelOpen(false)} onSubmit={handleCreate} loading={creating} />
    </div>
  );
}
