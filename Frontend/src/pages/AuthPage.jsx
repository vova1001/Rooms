import { ArrowLeft, ArrowRight, KeyRound, Mail, Sparkles } from 'lucide-react';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { sendEmailCode, verifyEmailCode } from '../api/http';
import Logo from '../components/Logo';

export default function AuthPage({ refreshSession }) {
  const navigate = useNavigate();
  const [step, setStep] = useState('email');
  const [email, setEmail] = useState('');
  const [code, setCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [message, setMessage] = useState('');

  async function submitEmail(event) {
    event.preventDefault();
    if (!email.trim() || loading) return;
    setLoading(true);
    setError('');
    setMessage('');
    try {
      await sendEmailCode(email);
      setStep('code');
      setMessage(`Код отправлен на ${email.trim()}`);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function submitCode(event) {
    event.preventDefault();
    if (!code.trim() || loading) return;
    setLoading(true);
    setError('');
    try {
      const result = await verifyEmailCode(email, code);
      const nextSession = await refreshSession();
      if (result?.requires_register || nextSession.status === 'registration') {
        navigate('/complete-profile', { replace: true });
      } else {
        navigate('/', { replace: true });
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function resend() {
    if (loading) return;
    setLoading(true);
    setError('');
    try {
      await sendEmailCode(email);
      setMessage('Новый код отправлен');
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="welcome-page auth-page">
      <header className="welcome-nav"><Logo /></header>
      <div className="welcome-glow" />
      <main className="welcome-content">
        <section className="welcome-copy">
          <span className="eyebrow"><Sparkles size={14} /> Пространство для голоса</span>
          <h1>Зайди.<br />Услышь своих.</h1>
          <p>Вход без пароля: отправим одноразовый код на почту и восстановим твою сессию автоматически.</p>
        </section>

        <section className="welcome-card auth-card">
          <span className="step-number">{step === 'email' ? '01' : '02'}</span>

          {step === 'email' ? (
            <>
              <div className="auth-icon"><Mail size={22} /></div>
              <h2>Войти по почте</h2>
              <p>Укажи email. Если профиль уже существует — сразу войдёшь, если нет — создадим его после подтверждения.</p>
              <form onSubmit={submitEmail}>
                <label htmlFor="email">Email</label>
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(event) => setEmail(event.target.value)}
                  placeholder="name@example.com"
                  autoComplete="email"
                  autoFocus
                  required
                />
                {error && <div className="error-message">{error}</div>}
                <button className="primary-button full" disabled={!email.trim() || loading} type="submit">
                  {loading ? 'Отправляем…' : 'Получить код'} <ArrowRight size={18} />
                </button>
              </form>
            </>
          ) : (
            <>
              <button className="auth-back" type="button" onClick={() => { setStep('email'); setCode(''); setError(''); }}>
                <ArrowLeft size={16} /> Изменить email
              </button>
              <div className="auth-icon"><KeyRound size={22} /></div>
              <h2>Проверь почту</h2>
              <p>Введи шестизначный код, который мы отправили на <strong>{email}</strong>.</p>
              <form onSubmit={submitCode}>
                <label htmlFor="code">Код подтверждения</label>
                <input
                  id="code"
                  className="code-input"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  value={code}
                  onChange={(event) => setCode(event.target.value.replace(/\D/g, '').slice(0, 6))}
                  placeholder="000000"
                  autoFocus
                />
                {message && <div className="success-message">{message}</div>}
                {error && <div className="error-message">{error}</div>}
                <button className="primary-button full" disabled={code.length !== 6 || loading} type="submit">
                  {loading ? 'Проверяем…' : 'Подтвердить'} <ArrowRight size={18} />
                </button>
                <button className="text-button" type="button" disabled={loading} onClick={resend}>Отправить код ещё раз</button>
              </form>
            </>
          )}
        </section>
      </main>
    </div>
  );
}
