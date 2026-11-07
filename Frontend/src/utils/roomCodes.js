function storageKey(room) {
  return `room-access-code:${room.ownerId}:${room.id}`;
}

export function readRoomCode(room) {
  if (!room.ownerId || !room.id) return '';
  try { return localStorage.getItem(storageKey(room)) || ''; }
  catch { return ''; }
}

export function storeRoomCode(room, isPrivate, password) {
  if (!room.ownerId || !room.id) return false;
  try {
    if (isPrivate) localStorage.setItem(storageKey(room), password);
    return true;
  } catch { return false; }
}
