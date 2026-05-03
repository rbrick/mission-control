import { Badge, Button, Card } from './ui';
import type { Rig } from '../lib/types';

export function RigSelector({ rigs, selectedRig, onSelect }: { rigs: Rig[]; selectedRig?: Rig; onSelect: (id: string) => void }) {
  return (
    <Card className="rig-strip">
      <div className="section-heading"><h2>Rigs</h2><p>{rigs.length} registered</p></div>
      <div className="rig-list">
        {rigs.length === 0 && <p className="muted">No rigs connected.</p>}
        {rigs.map((rig) => (
          <Button key={rig.id} variant="secondary" className={`rig-pill ${selectedRig?.id === rig.id ? 'selected' : ''}`} onClick={() => onSelect(rig.id)}>
            <span className={rig.online ? 'dot online' : 'dot'} />
            <span>{rig.id}</span>
            <Badge tone={rig.online ? 'green' : 'neutral'}>{rig.online ? 'online' : 'offline'}</Badge>
          </Button>
        ))}
      </div>
    </Card>
  );
}
