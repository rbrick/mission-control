import React, { useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { DndContext, DragEndEvent, PointerSensor, closestCenter, useSensor, useSensors } from '@dnd-kit/core';
import { SortableContext, arrayMove, rectSortingStrategy, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Activity, Bot, Camera, Circle, GripVertical, Maximize2, Minimize2, Plus, RefreshCw, SatelliteDish, Send, Telescope, X } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { Textarea } from '@/components/ui/textarea';
import { api, apiBase } from '@/lib/api';
import type { CommandState, Rig } from '@/lib/types';
import { cn } from '@/lib/utils';
import './styles.css';

type WidgetID = 'rig' | 'image' | 'guide' | 'chat' | 'controls' | 'state' | 'capabilities' | 'commands';
type WidgetSize = 'sm' | 'md' | 'lg';
type Layout = { order: WidgetID[]; sizes: Record<WidgetID, WidgetSize> };

const widgets: { id: WidgetID; title: string; defaultSize: WidgetSize }[] = [
  { id: 'rig', title: 'Rig', defaultSize: 'sm' },
  { id: 'image', title: 'Image', defaultSize: 'lg' },
  { id: 'guide', title: 'Guide graph', defaultSize: 'md' },
  { id: 'chat', title: 'Agent chat', defaultSize: 'md' },
  { id: 'controls', title: 'Controls', defaultSize: 'md' },
  { id: 'state', title: 'State', defaultSize: 'md' },
  { id: 'capabilities', title: 'Capabilities', defaultSize: 'md' },
  { id: 'commands', title: 'Command history', defaultSize: 'lg' },
];
const defaultOrder = widgets.map((w) => w.id);
const defaultSizes = Object.fromEntries(widgets.map((w) => [w.id, w.defaultSize])) as Record<WidgetID, WidgetSize>;
const titles = Object.fromEntries(widgets.map((w) => [w.id, w.title])) as Record<WidgetID, string>;

const quickCommands = [
  { namespace: 'rig', command: 'get_status' },
  { namespace: 'mount', command: 'goto_radec', data: { ra_hours: 10.684, dec_degrees: 41.269, epoch: 'J2000' } },
  { namespace: 'mount', command: 'park' },
  { namespace: 'camera', command: 'capture' },
  { namespace: 'sequence', command: 'start' },
  { namespace: 'sequence', command: 'stop' },
];

function App() {
  const [rigs, setRigs] = useState<Rig[]>([]);
  const [commands, setCommands] = useState<CommandState[]>([]);
  const [selectedRigID, setSelectedRigID] = useState('');
  const [error, setError] = useState('');
  const selectedRig = useMemo(() => rigs.find((rig) => rig.id === selectedRigID) ?? rigs[0], [rigs, selectedRigID]);

  async function refresh() {
    try {
      const [rigResponse, commandResponse] = await Promise.all([api<{ rigs: Rig[] }>('/v1/rigs'), api<{ commands: CommandState[] }>('/v1/commands')]);
      setRigs(rigResponse.rigs);
      setCommands([...commandResponse.commands].sort((a, b) => b.updated_at.localeCompare(a.updated_at)));
      if (!selectedRigID && rigResponse.rigs.length > 0) setSelectedRigID(rigResponse.rigs[0].id);
      setError('');
    } catch (err) { setError(err instanceof Error ? err.message : String(err)); }
  }
  useEffect(() => { refresh(); const id = window.setInterval(refresh, 1000); return () => window.clearInterval(id); }, []);

  async function sendCommand(namespace: string, command: string, data?: unknown) {
    if (!selectedRig) return;
    try { await api(`/v1/rigs/${selectedRig.id}/commands`, { method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ namespace, command, data }) }); await refresh(); }
    catch (err) { setError(err instanceof Error ? err.message : String(err)); }
  }

  return <main className="grid min-h-screen grid-cols-[280px_1fr] bg-white text-neutral-950 max-lg:grid-cols-1"><Sidebar rigs={rigs} selectedRig={selectedRig} onSelect={setSelectedRigID} /><section className="min-w-0 bg-gradient-to-b from-white to-neutral-50/80 px-8 py-7 max-md:px-4"><header className="mb-5 flex items-start justify-between gap-4"><div><p className="text-xs font-medium text-muted-foreground">{apiBase || 'Vite proxy /v1'}</p><h1 className="text-3xl font-semibold tracking-tight">{selectedRig?.id ?? 'Dashboard'}</h1></div><Button variant="outline" onClick={refresh}><RefreshCw className="size-4" />Refresh</Button></header>{error && <div className="mb-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}<br /><span className="text-red-600">Make sure the gateway is running. For non-8080 ports use VITE_GATEWAY_URL.</span></div>}<Dashboard rig={selectedRig} commands={commands.filter((command) => !selectedRig || command.rig_id === selectedRig.id)} onCommand={sendCommand} /></section></main>;
}

function Sidebar({ rigs, selectedRig, onSelect }: { rigs: Rig[]; selectedRig?: Rig; onSelect: (id: string) => void }) {
  return <aside className="sticky top-0 flex h-screen flex-col border-r bg-neutral-50/80 p-3 max-lg:static max-lg:h-auto max-lg:border-b max-lg:border-r-0"><div className="mb-5 flex items-center gap-3 rounded-xl px-2 py-2"><div className="grid size-9 place-items-center rounded-lg bg-neutral-950 text-xs font-bold text-white">MC</div><div><div className="font-semibold">Mission Control</div><div className="text-xs text-muted-foreground">Select a rig</div></div></div><div className="mb-2 px-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">Rigs</div><ScrollArea className="min-h-0 flex-1"><div className="grid gap-1 pr-2">{rigs.length === 0 && <p className="px-2 py-3 text-sm text-muted-foreground">No rigs connected.</p>}{rigs.map((rig) => <Button key={rig.id} variant="ghost" className={cn('h-auto justify-start gap-3 rounded-xl px-2 py-3 text-left', selectedRig?.id === rig.id && 'bg-accent')} onClick={() => onSelect(rig.id)}><Circle className={cn('size-2 fill-muted-foreground text-muted-foreground', rig.online && 'fill-emerald-500 text-emerald-500')} /><span className="grid flex-1"><span className="font-medium">{rig.id}</span><span className="text-xs text-muted-foreground">{rig.adapter ?? 'unknown adapter'}</span></span><Badge variant={rig.online ? 'default' : 'secondary'}>{rig.online ? 'on' : 'off'}</Badge></Button>)}</div></ScrollArea><div className="mt-3 rounded-xl border bg-white p-3 text-xs text-muted-foreground">Drag, resize, add, and remove widgets. Layout is saved per rig.</div></aside>;
}

function Dashboard({ rig, commands, onCommand }: { rig?: Rig; commands: CommandState[]; onCommand: (namespace: string, command: string, data?: unknown) => void }) {
  const storageKey = `mission-control.layout.${rig?.id ?? 'default'}`;
  const [layout, setLayout] = useState<Layout>(() => loadLayout(storageKey));
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 8 } }));
  const hidden = defaultOrder.filter((id) => !layout.order.includes(id));
  useEffect(() => setLayout(loadLayout(storageKey)), [storageKey]);
  useEffect(() => localStorage.setItem(storageKey, JSON.stringify(layout)), [layout, storageKey]);

  function onDragEnd(event: DragEndEvent) { const { active, over } = event; if (!over || active.id === over.id) return; setLayout((l) => ({ ...l, order: arrayMove(l.order, l.order.indexOf(active.id as WidgetID), l.order.indexOf(over.id as WidgetID)) })); }
  const remove = (id: WidgetID) => setLayout((l) => ({ ...l, order: l.order.filter((item) => item !== id) }));
  const add = (id: WidgetID) => setLayout((l) => ({ ...l, order: [...l.order, id] }));
  const resize = (id: WidgetID, size: WidgetSize) => setLayout((l) => ({ ...l, sizes: { ...l.sizes, [id]: size } }));
  const reset = () => setLayout({ order: defaultOrder, sizes: defaultSizes });

  return <><div className="mb-4 flex flex-wrap items-center gap-2 rounded-2xl border bg-white p-3 shadow-sm"><span className="mr-1 text-sm font-medium">Widgets</span>{hidden.map((id) => <Button key={id} size="sm" variant="outline" onClick={() => add(id)}><Plus className="size-3" />{titles[id]}</Button>)}{hidden.length === 0 && <span className="text-sm text-muted-foreground">All widgets visible</span>}<Button className="ml-auto" size="sm" variant="ghost" onClick={reset}>Reset layout</Button></div><DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={onDragEnd}><SortableContext items={layout.order} strategy={rectSortingStrategy}><div className="grid grid-cols-12 gap-4">{layout.order.map((id) => <DashboardWidget key={id} id={id} size={layout.sizes[id] ?? defaultSizes[id]} rig={rig} commands={commands} onCommand={onCommand} onRemove={remove} onResize={resize} />)}</div></SortableContext></DndContext></>;
}

function DashboardWidget(props: { id: WidgetID; size: WidgetSize; rig?: Rig; commands: CommandState[]; onCommand: (namespace: string, command: string, data?: unknown) => void; onRemove: (id: WidgetID) => void; onResize: (id: WidgetID, size: WidgetSize) => void }) {
  const { id, size, rig, commands, onCommand, onRemove, onResize } = props;
  switch (id) {
    case 'rig': return <SortableWidget {...props} title="Rig" icon={<Telescope className="size-4" />}><div className="flex items-start justify-between gap-4"><div><p className="text-sm text-muted-foreground">Selected rig</p><h2 className="mt-1 text-2xl font-semibold tracking-tight">{rig?.id ?? 'No rig selected'}</h2><p className="text-sm text-muted-foreground">{rig?.adapter ?? 'Waiting for connection'}</p></div><Badge variant={rig?.online ? 'default' : 'secondary'}>{rig?.online ? 'Online' : 'Offline'}</Badge></div></SortableWidget>;
    case 'image': return <SortableWidget {...props} title="Image" icon={<Camera className="size-4" />}><div className="grid min-h-[285px] place-items-center rounded-xl border border-dashed bg-neutral-50"><Empty title="Image preview" text="Live captures, solve overlays, HFR, and sequence progress." /></div></SortableWidget>;
    case 'guide': return <SortableWidget {...props} title="Guide graph" icon={<Activity className="size-4" />}><div className="relative grid min-h-[285px] place-items-center overflow-hidden rounded-xl border border-dashed bg-[linear-gradient(#eef0f3_1px,transparent_1px),linear-gradient(90deg,#eef0f3_1px,transparent_1px)] bg-[size:28px_28px]"><div className="absolute left-6 right-6 top-[42%] h-0.5 rotate-[-4deg] rounded-full bg-blue-600" /><div className="absolute left-6 right-6 top-[55%] h-0.5 rotate-[3deg] rounded-full bg-red-600" /><Empty title="PHD2 guide graph" text="RMS, RA/Dec corrections, dithers, and guide status." /></div></SortableWidget>;
    case 'chat': return <SortableWidget {...props} title="Agent chat" icon={<Bot className="size-4" />}><div className="flex min-h-[335px] flex-col gap-3"><div className="max-w-[82%] rounded-2xl bg-muted px-4 py-3 text-sm">How can I help run this rig tonight?</div><div className="ml-auto max-w-[82%] rounded-2xl bg-neutral-950 px-4 py-3 text-sm text-white">Agentic orchestration will connect here.</div><div className="mt-auto flex gap-2"><Textarea className="min-h-12 resize-none rounded-2xl" placeholder="Ask Mission Control..." /><Button size="icon"><Send className="size-4" /></Button></div></div></SortableWidget>;
    case 'controls': return <SortableWidget {...props} title="Controls" icon={<SatelliteDish className="size-4" />}><div className="grid grid-cols-2 gap-2 max-sm:grid-cols-1">{quickCommands.map((item) => <Button key={`${item.namespace}.${item.command}`} variant="outline" disabled={!rig?.online} onClick={() => onCommand(item.namespace, item.command, item.data)}>{item.namespace}.{item.command}</Button>)}</div></SortableWidget>;
    case 'state': return <SortableWidget {...props} title="State" icon={<Activity className="size-4" />}><pre className="max-h-48 overflow-auto rounded-lg bg-neutral-50 p-3 text-xs">{JSON.stringify(rig?.state ?? {}, null, 2)}</pre></SortableWidget>;
    case 'capabilities': return <SortableWidget {...props} title="Capabilities" icon={<Telescope className="size-4" />}>{rig?.capabilities?.length ? <div className="grid gap-2">{rig.capabilities.map((cap) => <div className="flex items-center gap-2 text-sm" key={`${cap.adapter_id}:${cap.namespace}`}><Badge variant="secondary">{cap.namespace}</Badge><span className="text-muted-foreground">{cap.commands.join(', ')}</span></div>)}</div> : <Empty title="No capabilities" text="Connect a rig to advertise commands." />}</SortableWidget>;
    case 'commands': return <SortableWidget {...props} title="Command history" icon={<Activity className="size-4" />}><ScrollArea className="h-[390px]"><div className="grid gap-2 pr-3">{commands.length === 0 && <Empty title="No commands yet" text="Send a command to see progress and results." />}{commands.slice(0, 12).map((cmd) => <div className="rounded-xl border p-3" key={cmd.id}><div className="flex items-center justify-between gap-3"><div><div className="font-medium">{cmd.namespace}.{cmd.command}</div><code className="text-xs text-muted-foreground">{cmd.id}</code></div><Badge variant={cmd.phase === 'error' ? 'destructive' : cmd.phase === 'result' ? 'default' : 'secondary'}>{cmd.phase}</Badge></div><Separator className="my-2" /><pre className="whitespace-pre-wrap text-xs text-muted-foreground">{JSON.stringify(cmd.error ?? cmd.data ?? {}, null, 2)}</pre></div>)}</div></ScrollArea></SortableWidget>;
  }
}

function SortableWidget({ id, size, title, icon, onRemove, onResize, children }: { id: WidgetID; size: WidgetSize; title: string; icon: React.ReactNode; onRemove: (id: WidgetID) => void; onResize: (id: WidgetID, size: WidgetSize) => void; children: React.ReactNode }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id });
  const style = { transform: CSS.Transform.toString(transform), transition };
  return <Card ref={setNodeRef} style={style} className={cn('group rounded-2xl shadow-sm max-xl:col-span-12', sizeClass(size), isDragging && 'z-50 opacity-80 ring-2 ring-ring')}><CardHeader className="flex flex-row items-center gap-2 space-y-0 pb-3"><button className="cursor-grab rounded-md p-1 text-muted-foreground opacity-60 hover:bg-muted hover:opacity-100 active:cursor-grabbing" {...attributes} {...listeners} aria-label={`Drag ${title} widget`}><GripVertical className="size-4" /></button>{icon}<CardTitle className="text-base">{title}</CardTitle><div className="ml-auto flex items-center gap-1 opacity-70 transition-opacity group-hover:opacity-100"><Button size="icon" variant="ghost" className="size-7" onClick={() => onResize(id, previousSize(size))}><Minimize2 className="size-3.5" /></Button><Button size="icon" variant="ghost" className="size-7" onClick={() => onResize(id, nextSize(size))}><Maximize2 className="size-3.5" /></Button><Button size="icon" variant="ghost" className="size-7" onClick={() => onRemove(id)}><X className="size-3.5" /></Button></div></CardHeader><CardContent>{children}</CardContent></Card>;
}

function Empty({ title, text }: { title: string; text: string }) { return <div className="grid place-items-center text-center"><div className="mb-2 grid size-10 place-items-center rounded-xl bg-muted">✦</div><div className="font-medium">{title}</div><p className="max-w-xs text-sm text-muted-foreground">{text}</p></div>; }
function sizeClass(size: WidgetSize) { return size === 'sm' ? 'col-span-4' : size === 'md' ? 'col-span-6' : 'col-span-8'; }
function nextSize(size: WidgetSize): WidgetSize { return size === 'sm' ? 'md' : size === 'md' ? 'lg' : 'lg'; }
function previousSize(size: WidgetSize): WidgetSize { return size === 'lg' ? 'md' : size === 'md' ? 'sm' : 'sm'; }
function loadLayout(storageKey: string): Layout { try { const parsed = JSON.parse(localStorage.getItem(storageKey) ?? 'null') as Partial<Layout> | null; if (parsed?.order) { const order = parsed.order.filter((id): id is WidgetID => defaultOrder.includes(id as WidgetID)); return { order, sizes: { ...defaultSizes, ...parsed.sizes } }; } } catch { /* ignore */ } return { order: defaultOrder, sizes: defaultSizes }; }

createRoot(document.getElementById('root')!).render(<App />);
