import { ArrowRight, Sparkles } from 'lucide-react';
import { useState } from 'react';
import { Navigate, useNavigate } from 'react-router-dom';
import { createUser } from '../api/http';
import Logo from '../components/Logo';
import { getStoredUser, saveUser } from '../utils/session';

export default function CreateUserPage() {
  const navigate = useNavigate();
  const existingUser = getStoredUser();
  const [username, setUsername] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  if (existingUser) return <Navigate to="/" replace />;

  async function submit(event) {
    event.preventDefault();
    if (loading) return;
    setLoading(true);
    setError('');
    try {
      const user = await createUser(username);
      saveUser(user);
      navigate('/', { replace: true });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="welcome-page">
      <header className="welcome-nav"><Logo /></header>
      <div className="welcome-glow" />
      <main className="welcome-content">
        <section className="welcome-copy">
          <span className="eyebrow"><Sparkles size={14} /> Пространство для голоса</span>
          <h1>Зайди.<br />Услышь своих.</h1>
          <p>Минималистичные голосовые комнаты без сложной регистрации и лишнего шума.</p>
        </section>
        <section className="welcome-card">
          <span className="step-number">01</span>
          <h2>Как тебя называть?</h2>
          <p>Создадим локальный профиль и сразу перенесём тебя к комнатам.</p>
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
            {error && <div className="error-message">{error}</div>}
            <button className="primary-button full" disabled={loading} type="submit">
              {loading ? 'Создаём профиль…' : 'Продолжить'} <ArrowRight size={18} />
            </button>
          </form>
        </section>
      </main>
    </div>
  );
}
