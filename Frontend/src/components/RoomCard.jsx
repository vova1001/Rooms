import { ArrowUpRight, Radio } from 'lucide-react';
import Avatar from './Avatar';

export default function RoomCard({ room, onJoin }) {
  const visibleUsers = room.users.slice(0, 4);
  const extra = Math.max(0, room.users.length - visibleUsers.length);

  return (
    <button className="room-card" onClick={() => onJoin(room)} type="button">
      <div className="room-card-top">
        <span className="live-dot"><Radio size={14} /> live</span>
        <ArrowUpRight size={20} className="room-arrow" />
      </div>
      <div>
        <h3>{room.name}</h3>
        <p>{room.users.length ? `${room.users.length} в эфире` : 'Пока никого'}</p>
      </div>
      <div className="avatar-stack" aria-label="Участники комнаты">
        {visibleUsers.map((user) => (
          <Avatar key={user.id} name={user.username} src={user.avatar} size="sm" />
        ))}
        {extra > 0 && <span className="avatar avatar-sm avatar-more">+{extra}</span>}
        {!visibleUsers.length && <span className="empty-members">Зайти первым</span>}
      </div>
    </button>
  );
}
