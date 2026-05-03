import { Badge, Button } from './ui';
import type { Rig } from '../lib/types';

export function Sidebar({ rigs, selectedRig, onSelect }: { rigs: Rig[]; selectedRig?: Rig; onSelect: (id: string) => void }) {
  return (
    <aside className="sidebar">
      <div className="sidebar-brand"><div className="logo">MC</div><div><strong>Mission Control</strong><span>Rigs</span></div></div>
      <div className="sidebar-section-title">Available rigs</div>
      <nav className="rig-nav">
        {rigs.length === 0 && <p className="muted empty-sidebar">No rigs connected.</p>}
        {rigs.map((rig) => (
          <Button key={rig.id} variant="ghost" className={`rig-nav-item ${selectedRig?.id === rig.id ? 'active' : ''}`} onClick={() => onSelect(rig.id)}>
            <span className={rig.online ? 'dot online' : 'dot'} />
            <span className="rig-nav-copy"><strong>{rig.id}</strong><small>{rig.adapter ?? 'unknown adapter'}</small></span>
            <Badge tone={rig.online ? 'green' : 'neutral'}>{rig.online ? 'on' : 'off'}</Badge>
          </Button>
        ))}
      </nav>
      <div className="sidebar-footer">Select a rig to control its dashboard workspace.</div>
    </aside>
  );
}
