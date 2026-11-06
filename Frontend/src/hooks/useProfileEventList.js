import { useCallback, useEffect, useRef, useState } from 'react';

export default function useProfileEventList(load, refreshSession) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const active = useRef(false);
  const version = useRef(0);
  const reload = useCallback(async () => {
    const requestVersion = ++version.current;
    setLoading(true);
    setError('');
    try {
      const data = await load();
      if (active.current && version.current === requestVersion) setItems(data);
    } catch (err) {
      if (!active.current || version.current !== requestVersion) return;
      if (err.status === 401) await refreshSession();
      if (active.current && version.current === requestVersion) setError('Не удалось загрузить данные. Попробуйте ещё раз.');
    } finally { if (active.current && version.current === requestVersion) setLoading(false); }
  }, [load, refreshSession]);
  useEffect(() => {
    active.current = true;
    reload();
    const update = () => { if (document.visibilityState === 'visible') reload(); };
    window.addEventListener('focus', update);
    document.addEventListener('visibilitychange', update);
    const timer = window.setInterval(update, 30000);
    return () => { active.current = false; version.current++; window.removeEventListener('focus', update); document.removeEventListener('visibilitychange', update); window.clearInterval(timer); };
  }, [reload]);
  function updateItems(update) {
    version.current++;
    setLoading(false);
    setItems(update);
  }
  return { items, loading, error, reload, updateItems };
}
