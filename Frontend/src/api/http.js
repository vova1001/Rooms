const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  const contentType = response.headers.get('content-type') || '';
  const data = contentType.includes('application/json')
    ? await response.json()
    : await response.text();

  if (!response.ok) {
    const message = typeof data === 'string'
      ? data
      : data?.message || data?.error || 'Ошибка запроса';
    throw new Error(message);
  }

  return data;
}

export function createUser(username) {
  return request('/user/init', {
    method: 'POST',
    body: JSON.stringify({ username: username.trim() }),
  });
}

export function getRooms() {
  return request('/rooms');
}

export function createRoom(name, userId) {
  return request('/rooms', {
    method: 'POST',
    headers: { 'X-User-ID': userId },
    body: JSON.stringify({ name: name.trim() }),
  });
}
