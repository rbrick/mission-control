export type Capability = {
  adapter_id?: string;
  adapter?: string;
  namespace: string;
  commands: string[];
};

export type Rig = {
  id: string;
  online: boolean;
  adapter?: string;
  capabilities?: Capability[];
  state?: unknown;
  last_seen?: string;
  connected_at?: string;
};

export type CommandState = {
  id: string;
  rig_id: string;
  namespace: string;
  command: string;
  phase: string;
  data?: unknown;
  error?: { code: string; message: string };
  updated_at: string;
};

export type Page = 'dashboard' | 'imaging' | 'guiding' | 'agent';
