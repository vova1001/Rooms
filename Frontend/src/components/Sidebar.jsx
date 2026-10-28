import { useNavigate } from 'react-router-dom';
import Avatar from './Avatar';
import Logo from './Logo';
import { getStoredUser } from '../utils/session';

export default function Sidebar({ active = '' }) {
  const navigate = useNavigate();
  const user = getStoredUser();

  return (
    <aside className="sidebar">
      <button className="sidebar-logo" onClick={() => navigate('/')} type="button" aria-label="К комнатам">
        <Logo />
      </button>
      <div className="sidebar-spacer" />
      <button
        className={`profile-block ${active === 'profile' ? 'active' : ''}`}
        onClick={() => navigate('/profile')}
        type="button"
        aria-label="Открыть профиль"
      >
        <Avatar name={user?.username} src={user?.avatar} />
        <div><strong>{user?.username || 'Пользователь'}</strong><span>В сети</span></div>
      </button>
      <div className="session-note">Вход защищён серверной сессией</div>
    </aside>
  );
}
