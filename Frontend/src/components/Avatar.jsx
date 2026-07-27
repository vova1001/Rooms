export default function Avatar({ name = '?', src = '', size = 'md' }) {
  const initials = name.trim().slice(0, 2).toUpperCase();
  return src ? (
    <img className={`avatar avatar-${size}`} src={src} alt={name} />
  ) : (
    <span className={`avatar avatar-${size} avatar-fallback`}>{initials}</span>
  );
}
