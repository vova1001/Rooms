import { ArrowRight, Image, Sparkles, UserRound } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { completeProfile, getAvatars, resolveApiAsset } from '../api/http';
import Logo from '../components/Logo';

function normalizeAvatar(entry, index) {
  if (typeof entry === 'string') return { id: entry, url: entry, imageUrl: resolveApiAsset(entry), label: `Аватар ${index + 1}` };
  const url = entry?.url || entry?.URL || entry?.avatar || entry?.Avatar || '';
  const id = entry?.id || entry?.ID || url || String(index);
  return { id, url, imageUrl: resolveApiAsset(url), label: `Аватар ${index + 1}` };
}

export default function CreateUserPage({ session, refreshSession }) {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [avatars, setAvatars] = useState([]);
  const [selectedAvatar, setSelectedAvatar] = useState('');
  const [loading, setLoading] = useState(false);
  const [avatarsLoading, setAvatarsLoading] = useState(true);
  const [error, setError] = useState('');
  const [avatarError, setAvatarError] = useState('');

  useEffect(() => {
    let active = true;
    async function load() {
      setAvatarsLoading(true);
      setAvatarError('');
      try {
        const data = await getAvatars();
        if (!active) return;
        const normalized = (Array.isArray(data) ? data : [])
          .map(normalizeAvatar)
          .filter((avatar) => avatar.url);
        setAvatars(normalized);
        if (normalized[0]) {
          setSelectedAvatar(normalized[0].url);
        } else if (Array.isArray(data) && data.length > 0) {
          setAvatarError('Бэкенд вернул записи аватаров без публичных полей id/url. Проверь, что поля структуры Avatars экспортируемые: ID и URL.');
        }
      } catch (err) {
        if (!active) return;
        setAvatarError(err.status === 401
          ? 'Список аватаров сейчас закрыт auth-middleware на бэкенде. Профиль можно создать без аватара; после открытия GET /avatars выбор появится автоматически.'
          : err.message);
      } finally {
        if (active) setAvatarsLoading(false);
      }
    }
    load();
    return () => { active = false; };
  }, []);

  const email = useMemo(() => session?.email || '', [session]);

  async function submit(event) {
    event.preventDefault();
    if (loading) return;
    setLoading(true);
    setError('');
    try {
      await completeProfile(username, selectedAvatar);
      const next = await refreshSession();
      if (next.status !== 'authorized') throw new Error('Профиль создан, но auth-сессия не подтвердилась');
      navigate('/', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="welcome-page profile-setup-page">
      <header className="welcome-nav"><Logo /></header>
      <div className="welcome-glow" />
      <main className="welcome-content profile-setup-content">
        <section className="welcome-copy">
          <span className="eyebrow"><Sparkles size={14} /> Последний шаг</span>
          <h1>Сделай профиль<br />своим.</h1>
          <p>Почта уже подтверждена. Осталось выбрать имя и аватар — после этого сервер заменит регистрационную сессию на обычную.</p>
          {email && <div className="confirmed-email">{email}</div>}
        </section>

        <section className="welcome-card profile-card">
          <span className="step-number">03</span>
          <div className="auth-icon"><UserRound size={22} /></div>
          <h2>Твой профиль</h2>
          <p>Имя будет видно в комнатах. Аватар можно выбрать из подготовленного набора.</p>

          <form onSubmit={submit}>
            <label htmlFor="username">Имя пользователя</label>
            <input
              id="username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="Можно оставить пустым"
              maxLength={40}
              autoFocus
            />

            <div className="avatar-picker-heading">
              <label>Аватар</label>
              {avatarsLoading && <span>Загружаем…</span>}
            </div>

            {avatars.length > 0 ? (
              <div className="avatar-picker">
                {avatars.map((avatar) => (
                  <button
                    key={avatar.id}
                    className={`avatar-option ${selectedAvatar === avatar.url ? 'selected' : ''}`}
                    type="button"
                    onClick={() => setSelectedAvatar(avatar.url)}
                    aria-label={avatar.label}
                  >
                    <img src={avatar.imageUrl} alt="" />
                  </button>
                ))}
              </div>
            ) : !avatarsLoading ? (
              <button className={`avatar-empty-option ${selectedAvatar === '' ? 'selected' : ''}`} type="button" onClick={() => setSelectedAvatar('')}>
                <Image size={21} />
                <span><strong>Без аватара</strong><small>Можно продолжить сейчас</small></span>
              </button>
            ) : <div className="avatar-picker-skeleton"><div className="loader" /></div>}

            {avatarError && <div className="warning-message">{avatarError}</div>}
            {error && <div className="error-message">{error}</div>}

            <button className="primary-button full" disabled={loading} type="submit">
              {loading ? 'Создаём профиль…' : 'Завершить регистрацию'} <ArrowRight size={18} />
            </button>
          </form>
        </section>
      </main>
    </div>
  );
}
