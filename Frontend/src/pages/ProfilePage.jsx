import { ArrowLeft, CalendarDays, Mail, RefreshCw } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getProfile } from '../api/http';
import Avatar from '../components/Avatar';
import Sidebar from '../components/Sidebar';

function formatDate(value) {
  if (!value) return '—';
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? '—'
    : date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
}

export default function ProfilePage({ refreshSession }) {
  const navigate = useNavigate();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  async function loadProfile() {
    setLoading(true);
    setError('');
    try {
      setProfile(await getProfile());
    } catch (err) {
      if (err?.status === 401) {
        const next = await refreshSession();
        if (next.status !== 'authorized') {
          navigate(next.status === 'registration' ? '/complete-profile' : '/welcome', { replace: true });
          return;
        }
      }
      setError(err.message || 'Не удалось загрузить профиль');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadProfile();
  }, []);

  return (
    <div className="app-shell">
      <Sidebar active="profile" />

      <main className="home-main profile-main">
        <button className="profile-back" onClick={() => navigate('/')} type="button">
          <ArrowLeft size={17} /> К комнатам
        </button>

        <header className="home-header">
          <div>
            <span className="eyebrow">Личный кабинет</span>
            <h1>Профиль</h1>
          </div>
        </header>

        {loading && <div className="empty-state"><div className="loader" /><h2>Загружаем профиль</h2></div>}

        {!loading && error && (
          <div className="profile-error">
            <div className="error-banner">{error}</div>
            <button className="secondary-button" onClick={loadProfile} type="button">
              <RefreshCw size={16} /> Попробовать снова
            </button>
          </div>
        )}

        {!loading && !error && profile && (
          <section className="profile-card">
            <div className="profile-hero">
              <Avatar name={profile.username || '?'} src={profile.avatar} size="xl" />
              <div className="profile-hero-text">
                <h2>{profile.username || 'Без имени'}</h2>
                <span className="profile-meta"><Mail size={14} /> {profile.email || '—'}</span>
                <span className="profile-meta"><CalendarDays size={14} /> С нами с {formatDate(profile.created_at)}</span>
              </div>
            </div>

            <div className="profile-section">
              <h3>О себе</h3>
              {profile.bio
                ? <p className="profile-bio">{profile.bio}</p>
                : <p className="profile-bio empty">Пока ничего не написано.</p>}
            </div>
          </section>
        )}
      </main>
    </div>
  );
}
