import { ArrowLeft, CalendarDays, Check, Mail, PencilLine, RefreshCw, X } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getProfile, updateProfile } from '../api/http';
import Avatar from '../components/Avatar';
import Sidebar from '../components/Sidebar';
import { getStoredUser, saveUser } from '../utils/session';

// Те же правила, что на сервере: сервер всё равно проверит, а здесь —
// чтобы сразу подсказать пользователю, не отправляя запрос.
const USERNAME_MAX = 50;
const BIO_MAX = 500;
// Управляющие (\n, \t, ...) и невидимые символы форматирования (zero-width, RTL-override).
const FORBIDDEN_RE = /[\p{Cc}\p{Cf}]/u;
const FORBIDDEN_BIO_RE = /[\p{Cf}]|[^\P{Cc}\n]/u;

function charCount(value) {
  return [...value].length;
}

function validate(username, bio) {
  const name = username.trim();
  if (!name) return 'Имя не может быть пустым';
  if (charCount(name) > USERNAME_MAX) return `Имя — не больше ${USERNAME_MAX} символов`;
  if (FORBIDDEN_RE.test(name)) return 'В имени есть недопустимые символы';
  const text = bio.trim();
  if (charCount(text) > BIO_MAX) return `«О себе» — не больше ${BIO_MAX} символов`;
  if (FORBIDDEN_BIO_RE.test(text)) return 'В тексте «О себе» есть недопустимые символы';
  return '';
}

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

  const [editing, setEditing] = useState(false);
  const [username, setUsername] = useState('');
  const [bio, setBio] = useState('');
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState('');
  const [saved, setSaved] = useState(false);

  async function handleUnauthorized(err) {
    if (err?.status !== 401) return false;
    const next = await refreshSession();
    if (next.status !== 'authorized') {
      navigate(next.status === 'registration' ? '/complete-profile' : '/welcome', { replace: true });
      return true;
    }
    return false;
  }

  async function loadProfile() {
    setLoading(true);
    setError('');
    try {
      setProfile(await getProfile());
    } catch (err) {
      if (!(await handleUnauthorized(err))) setError(err.message || 'Не удалось загрузить профиль');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadProfile();
  }, []);

  useEffect(() => {
    if (!saved) return undefined;
    const timer = setTimeout(() => setSaved(false), 2500);
    return () => clearTimeout(timer);
  }, [saved]);

  function startEditing() {
    setUsername(profile.username || '');
    setBio(profile.bio || '');
    setFormError('');
    setSaved(false);
    setEditing(true);
  }

  function cancelEditing() {
    setEditing(false);
    setFormError('');
  }

  async function save(event) {
    event.preventDefault();
    if (saving) return;

    const validationError = validate(username, bio);
    if (validationError) {
      setFormError(validationError);
      return;
    }

    // Шлём только то, что поменялось.
    const fields = {};
    if (username.trim() !== profile.username) fields.username = username.trim();
    if (bio.trim() !== (profile.bio || '')) fields.bio = bio.trim();
    if (Object.keys(fields).length === 0) {
      setEditing(false);
      return;
    }

    setSaving(true);
    setFormError('');
    try {
      const updated = await updateProfile(fields);
      setProfile(updated);
      // Обновляем сохранённого пользователя, чтобы имя сменилось и в боковой панели.
      const stored = getStoredUser();
      if (stored) saveUser({ ...stored, username: updated.username, avatar: updated.avatar });
      setEditing(false);
      setSaved(true);
    } catch (err) {
      if (!(await handleUnauthorized(err))) {
        // Сервер отдаёт "validation error: ..." — показываем без технического префикса.
        setFormError((err.message || 'Не удалось сохранить').replace(/^validation error:\s*/i, ''));
      }
    } finally {
      setSaving(false);
    }
  }

  const usernameLen = charCount(username.trim());
  const bioLen = charCount(bio.trim());

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
            <div className="profile-cover" />

            <div className="profile-hero">
              <div className="profile-avatar-ring">
                <Avatar name={profile.username || '?'} src={profile.avatar} size="xl" />
              </div>

              {!editing && (
                <button className="secondary-button profile-edit-button" onClick={startEditing} type="button">
                  <PencilLine size={16} /> Редактировать
                </button>
              )}
            </div>

            <div className="profile-hero-text">
              <h2>{profile.username || 'Без имени'}</h2>
              <div className="profile-meta-row">
                <span className="profile-meta"><Mail size={14} /> {profile.email || '—'}</span>
                <span className="profile-meta"><CalendarDays size={14} /> С нами с {formatDate(profile.created_at)}</span>
              </div>
            </div>

            {saved && <div className="profile-saved"><Check size={16} /> Изменения сохранены</div>}

            {!editing ? (
              <div className="profile-section">
                <h3>О себе</h3>
                {profile.bio
                  ? <p className="profile-bio">{profile.bio}</p>
                  : (
                    <button className="profile-bio-empty" onClick={startEditing} type="button">
                      Расскажите немного о себе — это увидят другие участники.
                    </button>
                  )}
              </div>
            ) : (
              <form className="profile-form" onSubmit={save} noValidate>
                <div className="profile-field">
                  <div className="profile-field-head">
                    <label htmlFor="profile-username">Имя</label>
                    <span className={usernameLen > USERNAME_MAX ? 'is-over' : ''}>{usernameLen} / {USERNAME_MAX}</span>
                  </div>
                  <input
                    id="profile-username"
                    value={username}
                    onChange={(event) => setUsername(event.target.value)}
                    placeholder="Как вас называть"
                    autoComplete="nickname"
                    autoFocus
                  />
                </div>

                <div className="profile-field">
                  <div className="profile-field-head">
                    <label htmlFor="profile-bio">О себе</label>
                    <span className={bioLen > BIO_MAX ? 'is-over' : ''}>{bioLen} / {BIO_MAX}</span>
                  </div>
                  <textarea
                    id="profile-bio"
                    value={bio}
                    onChange={(event) => setBio(event.target.value)}
                    placeholder="Пара слов о себе, интересы, чем занимаетесь"
                    rows={5}
                  />
                </div>

                {formError && <div className="error-message">{formError}</div>}

                <div className="profile-form-actions">
                  <button className="secondary-button" onClick={cancelEditing} type="button" disabled={saving}>
                    <X size={16} /> Отмена
                  </button>
                  <button className="primary-button" type="submit" disabled={saving}>
                    <Check size={16} /> {saving ? 'Сохраняем…' : 'Сохранить'}
                  </button>
                </div>
              </form>
            )}
          </section>
        )}
      </main>
    </div>
  );
}
