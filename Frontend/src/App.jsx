import { useCallback, useEffect, useState } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { getSession } from './api/http';
import Logo from './components/Logo';
import AuthPage from './pages/AuthPage';
import CreateUserPage from './pages/CreateUserPage';
import HomePage from './pages/HomePage';
import RoomPage from './pages/RoomPage';
import { clearUser, saveUser } from './utils/session';

function BootScreen() {
  return (
    <div className="session-loading-page">
      <Logo />
      <div className="session-loading-content">
        <div className="loader" />
        <span>Проверяем сессию…</span>
      </div>
    </div>
  );
}

function ProtectedRoute({ session, children }) {
  if (session.status !== 'authorized') {
    return <Navigate to={session.status === 'registration' ? '/complete-profile' : '/welcome'} replace />;
  }
  return children;
}

export default function App() {
  const [session, setSession] = useState({ status: 'loading' });

  const refreshSession = useCallback(async () => {
    setSession((current) => ({ ...current, status: current.status === 'loading' ? 'loading' : current.status }));
    try {
      const next = await getSession();
      setSession(next);
      if (next.status === 'authorized' && next.user) saveUser(next.user);
      else clearUser();
      return next;
    } catch (error) {
      clearUser();
      const next = { status: 'unauthorized', error: error.message };
      setSession(next);
      return next;
    }
  }, []);

  useEffect(() => {
    refreshSession();
  }, [refreshSession]);

  if (session.status === 'loading') return <BootScreen />;

  return (
    <Routes>
      <Route
        path="/welcome"
        element={
          session.status === 'authorized'
            ? <Navigate to="/" replace />
            : session.status === 'registration'
              ? <Navigate to="/complete-profile" replace />
              : <AuthPage refreshSession={refreshSession} />
        }
      />
      <Route
        path="/complete-profile"
        element={
          session.status === 'authorized'
            ? <Navigate to="/" replace />
            : session.status === 'registration'
              ? <CreateUserPage session={session} refreshSession={refreshSession} />
              : <Navigate to="/welcome" replace />
        }
      />
      <Route
        path="/"
        element={
          <ProtectedRoute session={session}>
            <HomePage refreshSession={refreshSession} />
          </ProtectedRoute>
        }
      />
      <Route
        path="/rooms/:roomId"
        element={
          <ProtectedRoute session={session}>
            <RoomPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to={session.status === 'authorized' ? '/' : '/welcome'} replace />} />
    </Routes>
  );
}
