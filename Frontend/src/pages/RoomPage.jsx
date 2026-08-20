import { ArrowLeft } from 'lucide-react';
import { useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import Logo from '../components/Logo';
import RoomExperience from '../components/RoomExperience';
import { useGatewayJoin } from '../hooks/useGatewayJoin';
import { getStoredUser } from '../utils/session';
import { normalizeRoom } from '../utils/rooms';

export default function RoomPage() {
  const { roomId } = useParams();
  const navigate = useNavigate();
  const user = getStoredUser();
  const room = useMemo(() => {
    try {
      const value = sessionStorage.getItem(`room:${roomId}`);
      return value ? normalizeRoom(JSON.parse(value)) : { id: roomId, name: 'Голосовая комната', users: [] };
    } catch {
      return { id: roomId, name: 'Голосовая комната', users: [] };
    }
  }, [roomId]);

  const { connection, status, error } = useGatewayJoin({ roomId, user });
  const leave = () => navigate('/');

  if (connection) {
    return <RoomExperience room={room} user={user} connection={connection} onLeave={leave} />;
  }

  return (
    <div className="connecting-page">
      <header><Logo /></header>
      <button className="back-link" onClick={leave} type="button"><ArrowLeft size={18} /> Назад к комнатам</button>
      <main className="connecting-card">
        <div className="signal-animation"><span /><span /><span /><span /></div>
        <span className="eyebrow">{status === 'error' ? 'Ошибка подключения' : 'Подключение'}</span>
        <h1>{room.name}</h1>
        <p>{error || 'Связываемся с Gateway и готовим голосовой канал…'}</p>
        {error && <button className="secondary-button" onClick={() => window.location.reload()} type="button">Попробовать снова</button>}
      </main>
    </div>
  );
}
