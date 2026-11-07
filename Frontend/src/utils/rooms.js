export function generateRoomCode() {
  const alphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789';
  return Array.from(crypto.getRandomValues(new Uint8Array(8)), (value) => alphabet[value & 31]).join('');
}

function normalizeUser(entry) {
  const value = entry?.user_info || entry?.userInfo || entry?.UserInfo || entry?.user || entry;
  return {
    id: value?.id || value?.ID || '',
    username: value?.username || value?.Username || 'Без имени',
    avatar: value?.avatar || value?.Avatar || '',
    createdAt: value?.created_at || value?.CreatedAt || null,
  };
}

export function normalizeRoom(room) {
  const rawUsers = room?.users || room?.room_users || room?.RoomUsers || [];
  return {
    id: room?.id || room?.ID,
    name: room?.name || room?.room_name || room?.RoomName || 'Без названия',
    ownerId: room?.owner_id || room?.OwnerID || room?.ownerId,
    createdAt: room?.created_at || room?.CreatedAt || room?.createdAt,
    isPrivate: room?.is_private ?? room?.isPrivate ?? false,
    hasAccess: room?.has_access ?? room?.hasAccess ?? false,
    users: Array.isArray(rawUsers) ? rawUsers.map(normalizeUser).filter((user) => user.id) : [],
  };
}
