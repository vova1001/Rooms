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
    ownerId: room?.owner_id || room?.OwnerID,
    createdAt: room?.created_at || room?.CreatedAt,
    users: Array.isArray(rawUsers) ? rawUsers.map(normalizeUser).filter((user) => user.id) : [],
  };
}
