export const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export function resolveApiAsset(url = '') {
  if (!url) return '';
  if (/^https?:\/\//i.test(url)) return url;
  return `${API_URL}${url.startsWith('/') ? '' : '/'}${url}`;
}

export class ApiError extends Error {
  constructor(message, status, data) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
  }
}

async function request(path, options = {}) {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    credentials: 'include',
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
      ? data.trim() || `HTTP ${response.status}`
      : data?.message || data?.error || `HTTP ${response.status}`;
    throw new ApiError(message, response.status, data);
  }

  return data;
}

export async function getSession() {
  try {
    return await request('/auth/session');
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      return { status: 'unauthorized' };
    }
    throw error;
  }
}

export function sendEmailCode(email) {
  return request('/auth/email/send-code', {
    method: 'POST',
    body: JSON.stringify({ email: email.trim() }),
  });
}

export function verifyEmailCode(email, code) {
  return request('/auth/email/verify-code', {
    method: 'POST',
    body: JSON.stringify({ email: email.trim(), code: code.trim() }),
  });
}

export function getAvatars() {
  return request('/avatars');
}

export function unlockAvatars(code) {
  return request('/avatars/unlock', {
    method: 'POST',
    body: JSON.stringify({ code }),
  });
}

export function completeProfile(username, avatar) {
  return request('/create-user', {
    method: 'POST',
    body: JSON.stringify({
      username: username.trim(),
      avatar: avatar || '',
    }),
  });
}

export function getRooms() {
  return request('/rooms');
}

export function createRoom(name, userId) {
  return request('/create-rooms', {
    method: 'POST',
    headers: userId ? { 'X-User-ID': userId } : {},
    body: JSON.stringify({ name: name.trim() }),
  });
}
