import { AudioLines } from 'lucide-react';

export default function Logo({ compact = false }) {
  return (
    <div className="logo" aria-label="Hush">
      <span className="logo-mark"><AudioLines size={20} /></span>
      {!compact && <span>hush</span>}
    </div>
  );
}
