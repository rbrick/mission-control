import { Card, EmptyState } from './ui';

export function ImagingPage() {
  return <div className="workspace-grid"><Card className="large-placeholder"><EmptyState title="Image preview" description="Live captures, solve overlays, HFR, focus metrics, and sequence progress will live here." /></Card><Card><h2>Capture details</h2><p className="muted">Exposure, filter, target, plate-solve result, and file metadata.</p></Card></div>;
}

export function GuidingPage() {
  return <div className="workspace-grid"><Card className="large-placeholder graph"><EmptyState title="Guide graph" description="PHD2 RMS, RA/Dec corrections, dithers, star mass, and guide status." /></Card><Card><h2>Guiding status</h2><p className="muted">Reserved for guider connection and calibration details.</p></Card></div>;
}

export function AgentPage() {
  return <div className="chat-layout"><Card className="chat-card"><div className="message assistant">How can I help run the observatory tonight?</div><div className="message user muted">Agentic control will connect here after rig APIs settle.</div><div className="chat-input">Ask Mission Control...</div></Card><Card><h2>Context</h2><p className="muted">Future agent context: rigs, safety, weather, target plan, active sequence, and recovery actions.</p></Card></div>;
}
