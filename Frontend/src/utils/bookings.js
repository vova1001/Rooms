export function groupProfileBookings(bookings, now = Date.now()) {
  const groups = { pending: [], confirmed: [], rejected: [], past: [] };
  for (const booking of bookings) {
    const end = Date.parse(booking.event?.ends_at);
    const key = booking.status === 'confirmed' && Number.isFinite(end) && end <= now ? 'past' : booking.status;
    if (groups[key]) groups[key].push(booking);
  }
  groups.past.sort((a, b) => Date.parse(b.event.ends_at) - Date.parse(a.event.ends_at));
  groups.confirmed.sort((a, b) => Date.parse(a.event.starts_at) - Date.parse(b.event.starts_at));
  return groups;
}

export function groupOrganizerEvents(events, now = Date.now()) {
  const groups = { upcoming: [], past: [] };
  for (const event of events) {
    const end = Date.parse(event.ends_at);
    groups[Number.isFinite(end) && end <= now ? 'past' : 'upcoming'].push(event);
  }
  groups.upcoming.sort((a, b) => Date.parse(a.starts_at) - Date.parse(b.starts_at));
  groups.past.sort((a, b) => Date.parse(b.ends_at) - Date.parse(a.ends_at));
  return groups;
}
