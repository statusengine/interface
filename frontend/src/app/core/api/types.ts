/** Shapes the API returns. Kept in one file so a contract change is one diff. */

/** Every list endpoint answers with this envelope. */
export interface ListResponse<T> {
  data: T[];
  meta: ListMeta;
}

export interface ListMeta {
  total: number;
  limit: number;
  offset: number;
  sort?: string;
}

/** Every failure answers with this. `code` is stable and translatable;
 *  `message` is for a human reading a log. */
export interface ApiErrorBody {
  error: {
    code: string;
    message: string;
    field?: string;
  };
}

/** What the server tells an anonymous visitor about itself. */
export interface ServerMeta {
  product: string;
  version: string;
  demo_mode: boolean;
  commands_enabled: boolean;
  events_enabled: boolean;
  metrics_provider: string;
  default_page_size: number;
  max_page_size: number;
}

/** The signed-in user. `permissions` is already expanded, so the UI never
 *  has to know that the admin role is stored as a wildcard. */
export interface Identity {
  username: string;
  display_name: string;
  email: string;
  role: string;
  permissions: Permission[];
  is_demo: boolean;
  read_only: boolean;
  expires_at: number;
}

export type Permission =
  | 'hosts:read'
  | 'services:read'
  | 'problems:read'
  | 'downtimes:read'
  | 'acknowledgements:read'
  | 'logentries:read'
  | 'history:read'
  | 'metrics:read'
  | 'audit:read'
  | 'commands:reschedule'
  | 'commands:acknowledge'
  | 'commands:downtime'
  | 'commands:notification'
  | 'commands:passiveresult'
  | 'commands:toggle'
  | 'users:manage';
