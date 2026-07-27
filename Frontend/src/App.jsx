import { Navigate, Route, Routes } from 'react-router-dom';
import CreateUserPage from './pages/CreateUserPage';
import HomePage from './pages/HomePage';
import RoomPage from './pages/RoomPage';
import { getStoredUser } from './utils/session';

function ProtectedRoute({ children }) {
  return getStoredUser() ? children : <Navigate to="/welcome" replace />;
}

export default function App() {
  return (
    <Routes>
      <Route path="/welcome" element={<CreateUserPage />} />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <HomePage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/rooms/:roomId"
        element={
          <ProtectedRoute>
            <RoomPage />
          </ProtectedRoute>
        }
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
