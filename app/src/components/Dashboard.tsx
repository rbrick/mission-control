import { Badge, Button, Card, EmptyState } from './ui';
import type { CommandState, Rig } from '../lib/types';

const quickCommands = [
  { namespace: 'rig', command: 'get_status' },
  { namespace: 'mount', command: 'goto_radec', data: { ra_hours: 10.684, dec_degrees: 41.269, epoch: 'J2000' } },
  { namespace: 'mount', command: 'park' },
  { namespace: 'camera', command: 'capture' },
  { namespace: 'sequence', command: 'start' },
  { namespace: 'sequence', command: 'stop' },
];

export function Dashboard({ rig, commands, onCommand }: { rig?: Rig; commands: CommandState[]; onCommand: (namespace: string, command: string, data?: unknown) => void }) {
  return (
    <div className="workspace">
      <Card className="widget hero-widget draggable"><WidgetChrome title="Rig" /><div className="hero-body"><div><p className="eyebrow">Selected rig</p><h2>{rig?.id ?? 'No rig selected'}</h2><p className="muted">{rig?.adapter ?? 'Waiting for a rig connection'}</p></div><Badge tone={rig?.online ? 'green' : 'neutral'}>{rig?.online ? 'Online' : 'Offline'}</Badge></div></Card>
      <Card className="widget image-widget draggable"><WidgetChrome title="Image" /><div className="image-preview"><EmptyState title="Image preview" description="Live captures, solve overlays, HFR, and sequence progress." /></div></Card>
      <Card className="widget guide-widget draggable"><WidgetChrome title="Guide graph" /><div className="guide-graph"><div className="guide-line guide-ra" /><div className="guide-line guide-dec" /><EmptyState title="PHD2 guide graph" description="RMS, RA/Dec corrections, dithers, and guide status." /></div></Card>
      <Card className="widget chat-widget draggable"><WidgetChrome title="Agent chat" /><div className="chat-card"><div className="message assistant">How can I help run this rig tonight?</div><div className="message user muted">Agentic orchestration will connect here.</div><div className="chat-input">Ask Mission Control...</div></div></Card>
      <Card className="widget controls-widget draggable"><WidgetChrome title="Controls" /><div className="command-grid">{quickCommands.map((item) => <Button key={`${item.namespace}.${item.command}`} disabled={!rig?.online} onClick={() => onCommand(item.namespace, item.command, item.data)}>{item.namespace}.{item.command}</Button>)}</div></Card>
      <Card className="widget state-widget draggable"><WidgetChrome title="State" /><pre>{JSON.stringify(rig?.state ?? {}, null, 2)}</pre></Card>
      <Card className="widget capabilities-widget draggable"><WidgetChrome title="Capabilities" />{rig?.capabilities?.length ? <div className="cap-list">{rig.capabilities.map((cap) => <div className="cap-row" key={`${cap.adapter_id}:${cap.namespace}`}><Badge tone="blue">{cap.namespace}</Badge><span>{cap.commands.join(', ')}</span></div>)}</div> : <EmptyState title="No capabilities" description="Connect a rig to advertise commands." />}</Card>
      <CommandTable commands={commands} />
    </div>
  );
}

function WidgetChrome({ title }: { title: string }) {
  return <div className="widget-chrome"><span className="drag-handle">⋮⋮</span><h2>{title}</h2></div>;
}

export function CommandTable({ commands }: { commands: CommandState[] }) {
  return <Card className="widget commands-widget draggable"><WidgetChrome title="Command history" /><div className="command-list">{commands.length === 0 && <EmptyState title="No commands yet" description="Send a command to see progress and results." />}{commands.slice(0, 12).map((cmd) => <div className="command-row" key={cmd.id}><div><strong>{cmd.namespace}.{cmd.command}</strong><code>{cmd.id}</code></div><Badge tone={phaseTone(cmd.phase)}>{cmd.phase}</Badge><pre>{JSON.stringify(cmd.error ?? cmd.data ?? {}, null, 2)}</pre></div>)}</div></Card>;
}

function phaseTone(phase: string) {
  if (phase === 'result') return 'green';
  if (phase === 'error') return 'red';
  if (phase === 'progress') return 'blue';
  return 'violet';
}
