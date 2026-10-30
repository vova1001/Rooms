import { Link, useLocation, useNavigate } from 'react-router-dom';
import { CalendarDays, Radio } from 'lucide-react';
import Avatar from './Avatar';
import Logo from './Logo';
import { getStoredUser } from '../utils/session';

export default function Sidebar({ active = '' }) {
  const navigate = useNavigate();
  const location = useLocation();
  const eventsActive = location.pathname === '/' && new URLSearchParams(location.search).get('tab') === 'events';
  const user = getStoredUser();

  return (
    <aside className="sidebar">
      <button className="sidebar-logo" onClick={() => navigate('/')} type="button" aria-label="К комнатам">
        <Logo />
      </button>
      <div className="sidebar-spacer" />
      <nav className="sidebar-navigation" aria-label="Разделы приложения">
        <Link to="/" className={`sidebar-nav-link ${location.pathname === '/' && !eventsActive ? 'is-active' : ''}`} aria-current={location.pathname === '/' && !eventsActive ? 'page' : undefined}><Radio size={19} /><span>Комнаты</span></Link>
        <Link to="/?tab=events" className={`sidebar-nav-link ${eventsActive ? 'is-active' : ''}`} aria-current={eventsActive ? 'page' : undefined}><CalendarDays size={19} /><span>Мероприятия</span></Link>
      </nav>
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
