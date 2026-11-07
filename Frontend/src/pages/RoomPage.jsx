import { ArrowLeft, LockKeyhole } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { getRooms, unlockRoom } from '../api/http';
import Logo from '../components/Logo';
import RoomExperience from '../components/RoomExperience';
import { useGatewayJoin } from '../hooks/useGatewayJoin';
import { getStoredUser } from '../utils/session';
import { normalizeRoom } from '../utils/rooms';

function RoomLoadingScreen({ startedAt, onLeave, children }) {
  const [visible, setVisible] = useState(() => Date.now() - startedAt >= 250);
  useEffect(() => {
    const remaining = 250 - (Date.now() - startedAt);
    if (remaining <= 0) { setVisible(true); return; }
    setVisible(false);
    const timer = window.setTimeout(() => setVisible(true), remaining);
    return () => window.clearTimeout(timer);
  }, [startedAt]);
  if (!visible) return <div className="room-transition" aria-busy="true" aria-label="Загружаем комнату" />;
  return <RoomEntryLayout onLeave={onLeave} loading>{children}</RoomEntryLayout>;
}

function ConnectedRoom({ room, user, refreshSession, onLeave, onRetry, onPrivacyChange, onAccessRequired, startedAt }) {
  const { connection, status, error, chat } = useGatewayJoin({ roomId: room.id, user });
  useEffect(() => {
    if (!error || connection) return;
    let active = true;
    getRooms().then((rooms) => {
      const found = rooms.map(normalizeRoom).find((item) => item.id === room.id);
      if (active && found?.isPrivate && !found.hasAccess) onAccessRequired(found);
    }).catch(() => {});
    return () => { active = false; };
  }, [error, connection, room.id, onAccessRequired]);
  if (connection) return <RoomExperience room={room} user={user} connection={connection} chat={chat} onLeave={onLeave} refreshSession={refreshSession} onPrivacyChange={onPrivacyChange} />;
  const Layout = error || status === 'closed' ? RoomEntryLayout : RoomLoadingScreen;
  return <Layout onLeave={onLeave} startedAt={startedAt}>
    <div className="signal-animation"><span /><span /><span /></div>
    <span className="eyebrow">{status === 'error' ? 'Ошибка подключения' : 'Подключение'}</span>
    <h1>{room.name}</h1>
    <p>{error || 'Связываемся с Gateway и готовим голосовой канал…'}</p>
    {(error || status === 'closed') && <button className="secondary-button" onClick={onRetry} type="button">Проверить доступ и повторить</button>}
  </Layout>;
}

function RoomEntryLayout({ onLeave, children, loading = false }) {
  return <div className={`connecting-page${loading ? ' is-loading' : ''}`}><header><Logo /></header>
    <button className="back-link" onClick={onLeave} type="button"><ArrowLeft size={18} /> Назад к комнатам</button>
    <main className="connecting-card">{children}</main>
  </div>;
}

function RoomEntry({ roomId, refreshSession }) {
  const navigate = useNavigate();
  const user = getStoredUser();
  const [room, setRoom] = useState(null);
  const [stage, setStage] = useState('loading');
  const [error, setError] = useState('');
  const [password, setPassword] = useState('');
  const [saving, setSaving] = useState(false);
  const active = useRef(false);
  const inFlight = useRef(false);
  const version = useRef(0);
  const startedAt = useRef(Date.now());
  const requireCode = useCallback((found) => { setRoom(found); setStage('code'); setError('Для входа в эту комнату нужен код.'); }, []);
  const leave = () => navigate('/');
  const load = useCallback(async () => {
    startedAt.current = Date.now();
    const current = ++version.current;
    setStage('loading');
    setError('');
    try {
      const rooms = await getRooms();
      if (!active.current || current !== version.current) return;
      const found = rooms.map(normalizeRoom).find((item) => item.id === roomId);
      if (!found) { setStage('missing'); setError('Комната не найдена.'); return; }
      setRoom(found);
      setStage(found.isPrivate && !found.hasAccess ? 'code' : 'allowed');
    } catch (err) {
      if (!active.current || current !== version.current) return;
      if (err.status === 401) await refreshSession();
      if (active.current && current === version.current) { setStage('error'); setError('Не удалось проверить доступ к комнате. Попробуйте ещё раз.'); }
    }
  }, [roomId, refreshSession]);
  useEffect(() => {
    active.current = true;
    load();
    return () => { active.current = false; version.current++; };
  }, [load]);
  async function submit(event) {
    event.preventDefault();
    if (!password || inFlight.current) return;
    inFlight.current = true;
    setSaving(true);
    setError('');
    try {
      await unlockRoom(roomId, password);
      if (!active.current) return;
      startedAt.current = Date.now();
      setPassword('');
      setRoom((current) => ({ ...current, hasAccess: true }));
      setStage('allowed');
    } catch (err) {
      if (!active.current) return;
      if (err.status === 401) await refreshSession();
      if (!active.current) return;
      if (err.status === 404) setStage('missing');
      setError(err.status === 403 ? 'Неверный код комнаты.' : err.status === 429 ? 'Слишком много попыток. Подождите и попробуйте позже.' : err.status === 404 ? 'Комната не найдена.' : 'Не удалось проверить код. Попробуйте ещё раз.');
    } finally {
      inFlight.current = false;
      if (active.current) setSaving(false);
    }
  }
  if (stage === 'allowed') return <ConnectedRoom room={room} user={user} refreshSession={refreshSession} onLeave={leave} onRetry={load} onAccessRequired={requireCode} startedAt={startedAt.current} onPrivacyChange={(isPrivate) => setRoom((current) => ({ ...current, isPrivate }))} />;
  if (stage === 'loading') return <RoomLoadingScreen startedAt={startedAt.current} onLeave={leave}><div className="loader" /><h1>Проверяем доступ</h1></RoomLoadingScreen>;
  return <RoomEntryLayout onLeave={leave}>
    <>
      {stage === 'code' && <LockKeyhole size={28} />}
      <h1>{room?.name || 'Вход в комнату'}</h1>
      {stage === 'code' && <><p>Это закрытая комната. Введите код, который вам передал организатор.</p>
        <form className="room-unlock-form" onSubmit={submit}>
          <label htmlFor="room-access-code">Код комнаты</label>
          <input id="room-access-code" type="password" autoComplete="off" value={password} onChange={(event) => setPassword(event.target.value)} disabled={saving} required aria-describedby={error ? 'room-access-error' : undefined} />
          <button className="primary-button full" type="submit" disabled={saving || !password}>{saving ? 'Проверяем…' : 'Войти в комнату'}</button>
        </form></>}
      {error && <p id="room-access-error" className="error-message" role="alert">{error}</p>}
      {stage === 'error' && <button className="secondary-button" onClick={load} type="button">Попробовать снова</button>}
    </>
  </RoomEntryLayout>;
}

export default function RoomPage({ refreshSession }) {
  const { roomId } = useParams();
  return <RoomEntry key={roomId} roomId={roomId} refreshSession={refreshSession} />;
}
