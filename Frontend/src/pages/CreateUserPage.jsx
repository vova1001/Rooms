import { ArrowRight, Image, KeyRound, Sparkles, UserRound, X } from 'lucide-react';
import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { completeProfile, getAvatars, resolveApiAsset, unlockAvatars } from '../api/http';
import Logo from '../components/Logo';

const SECRET_AVATAR_COUNT = 6;

function isSecretAvatarUrl(url = '') {
  const clean = String(url).split('?')[0].split('#')[0];
  const fileName = clean.slice(clean.lastIndexOf('/') + 1).toLowerCase();
  return /^s[1-6]\.png$/.test(fileName);
}

function normalizeAvatar(entry, index) {
  if (typeof entry === 'string') {
    return {
      id: entry,
      url: entry,
      imageUrl: resolveApiAsset(entry),
      label: `Аватар ${index + 1}`,
      isSecret: false,
    };
  }

  const url = entry?.url || entry?.URL || entry?.avatar || entry?.Avatar || '';
  const id = entry?.id || entry?.ID || url || String(index);
  const explicitSecret = entry?.is_secret ?? entry?.isSecret ?? entry?.secret ?? entry?.Secret;
  const isSecret = explicitSecret == null ? isSecretAvatarUrl(url) : Boolean(explicitSecret) || isSecretAvatarUrl(url);

  return {
    id,
    url,
    imageUrl: resolveApiAsset(url),
    label: `Аватар ${index + 1}`,
    isSecret,
  };
}

function splitInitialAvatars(items) {
  const detectedSecrets = items.filter((avatar) => avatar.isSecret || isSecretAvatarUrl(avatar.url));

  if (detectedSecrets.length > 0) {
    return {
      publicAvatars: items.filter((avatar) => !(avatar.isSecret || isSecretAvatarUrl(avatar.url))),
      secretAvatars: detectedSecrets,
    };
  }

  // Fallback для старого backend без признака is_secret и без имён s1–s6.
  // Используем его только если никакие секретные файлы определить невозможно.
  if (items.length >= SECRET_AVATAR_COUNT) {
    return {
      publicAvatars: items.slice(0, -SECRET_AVATAR_COUNT),
      secretAvatars: items.slice(-SECRET_AVATAR_COUNT),
    };
  }

  return { publicAvatars: items, secretAvatars: [] };
}

export default function CreateUserPage({ session, refreshSession }) {
  const navigate = useNavigate();
  const [username, setUsername] = useState('');
  const [avatars, setAvatars] = useState([]);
  const [lockedAvatars, setLockedAvatars] = useState([]);
  const [selectedAvatar, setSelectedAvatar] = useState('');
  const [loading, setLoading] = useState(false);
  const [avatarsLoading, setAvatarsLoading] = useState(true);
  const [error, setError] = useState('');
  const [avatarError, setAvatarError] = useState('');
  const [unlockOpen, setUnlockOpen] = useState(false);
  const [unlockCode, setUnlockCode] = useState('');
  const [unlockLoading, setUnlockLoading] = useState(false);
  const [unlockError, setUnlockError] = useState('');
  const [secretUnlocked, setSecretUnlocked] = useState(false);

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

        const { publicAvatars, secretAvatars } = splitInitialAvatars(normalized);
        setAvatars(publicAvatars);
        setLockedAvatars(secretAvatars);

        if (publicAvatars[0]) {
          setSelectedAvatar(publicAvatars[0].url);
        } else if (normalized.length > 0) {
          setAvatarError('Бэкенд вернул записи аватаров без публичных полей id/url.');
        }
      } catch (err) {
        if (!active) return;
        setAvatarError(err.message);
      } finally {
        if (active) setAvatarsLoading(false);
      }
    }

    load();
    return () => { active = false; };
  }, []);

  const email = useMemo(() => session?.email || '', [session]);

  async function handleUnlock(event) {
    event.preventDefault();
    if (!unlockCode.trim() || unlockLoading) return;

    setUnlockLoading(true);
    setUnlockError('');

    try {
      const result = await unlockAvatars(unlockCode.trim());

      const returned = Array.isArray(result)
        ? result
        : Array.isArray(result?.avatars)
          ? result.avatars
          : Array.isArray(result?.data)
            ? result.data
            : [];

      // Ответ unlock может содержать и публичные, и секретные записи.
      // Берём из него только s1.png–s6.png / is_secret=true, а уже скрытый набор
      // из первоначального GET /avatars используем как основной источник.
      const returnedSecrets = returned
        .map((item, index) => normalizeAvatar(item, avatars.length + index))
        .filter((avatar) => avatar.url && (avatar.isSecret || isSecretAvatarUrl(avatar.url)));

      const secretMap = new Map();
      [...lockedAvatars, ...returnedSecrets].forEach((avatar) => {
        const key = avatar.url;
        if (key) secretMap.set(key, avatar);
      });

      let secrets = [...secretMap.values()];

      if (secrets.length === 0) {
        // Если unlock возвращает только {success:true}, перечитываем /avatars.
        const refreshed = await getAvatars();
        secrets = (Array.isArray(refreshed) ? refreshed : [])
          .map(normalizeAvatar)
          .filter((avatar) => avatar.url && (avatar.isSecret || isSecretAvatarUrl(avatar.url)));
      }

      setAvatars((current) => {
        const byUrl = new Map(current.map((avatar) => [avatar.url, avatar]));
        secrets.forEach((avatar) => byUrl.set(avatar.url, avatar));
        return [...byUrl.values()];
      });
      setLockedAvatars([]);
      setSecretUnlocked(true);
      setUnlockOpen(false);
      setUnlockCode('');
    } catch (err) {
      setUnlockError(err.message || 'Неверный код');
    } finally {
      setUnlockLoading(false);
    }
  }

  async function submit(event) {
    event.preventDefault();
    if (loading) return;
    setLoading(true);
    setError('');

    try {
      await completeProfile(username, selectedAvatar);
      const next = await refreshSession();
      if (next.status !== 'authorized') {
        throw new Error('Профиль создан, но auth-сессия не подтвердилась');
      }
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
              {!avatarsLoading && secretUnlocked && <span className="secret-unlocked-label">Секретные открыты</span>}
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
                    <span className="avatar-option-image">
                      <img src={avatar.imageUrl} alt="" />
                    </span>
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

      <div className={`secret-avatar-control ${unlockOpen ? 'open' : ''}`}>
        {unlockOpen && (
          <form className="secret-avatar-popover" onSubmit={handleUnlock}>
            <button
              className="secret-popover-close"
              type="button"
              onClick={() => {
                setUnlockOpen(false);
                setUnlockError('');
              }}
              aria-label="Закрыть"
            >
              <X size={16} />
            </button>
            <div className="secret-popover-icon"><KeyRound size={18} /></div>
            <strong>Секретный набор</strong>
            <span>Введи код, чтобы открыть дополнительные аватары.</span>
            <input
              value={unlockCode}
              onChange={(event) => setUnlockCode(event.target.value)}
              placeholder="Код доступа"
              autoComplete="off"
              autoFocus
            />
            {unlockError && <div className="secret-unlock-error">{unlockError}</div>}
            <button className="secret-unlock-submit" type="submit" disabled={unlockLoading || !unlockCode.trim()}>
              {unlockLoading ? 'Проверяем…' : 'Открыть'}
            </button>
          </form>
        )}

        {!secretUnlocked && (
          <button
            className="secret-avatar-trigger"
            type="button"
            onClick={() => {
              setUnlockOpen((value) => !value);
              setUnlockError('');
            }}
            aria-label="Секретные аватары"
          >
            ?
          </button>
        )}
      </div>
    </div>
  );
}
