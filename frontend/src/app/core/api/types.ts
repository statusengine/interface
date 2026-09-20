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
  /** What the public demo account may submit, empty when read-only. */
  demo_commands?: string[];
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

// --- monitoring ------------------------------------------------------------
//
// Timestamps are Unix seconds. Zero means "never" - the worker writes it
// for things that have not happened yet - so never render one as a date.

export type Kind = 'host' | 'service';

/** Fields both status tables share. */
interface StatusCommon {
  state: number;
  state_text: string;
  is_hard_state: boolean;

  output: string;
  long_output?: string;
  perfdata?: string;

  current_check_attempt: number;
  max_check_attempts: number;
  last_check: number;
  next_check: number;
  last_state_change: number;
  last_hard_state_change: number;
  status_update_time: number;

  acknowledged: boolean;
  acknowledgement_type: number;
  in_downtime: boolean;
  scheduled_downtime_depth: number;
  is_flapping: boolean;
  notifications_enabled: boolean;
  active_checks_enabled: boolean;
  passive_checks_enabled: boolean;
  is_passive_check: boolean;
  event_handler_enabled: boolean;
  flap_detection_enabled: boolean;

  latency: number;
  execution_time: number;

  // Detail only.
  check_command?: string;
  event_handler?: string;
  check_timeperiod?: string;
  node_name?: string;
  normal_check_interval?: number;
  retry_check_interval?: number;
  percent_state_change?: number;
  last_notification?: number;
  next_notification?: number;
  current_notification_number?: number;
}

export interface HostStatus extends StatusCommon {
  hostname: string;
  last_time_up?: number;
  last_time_down?: number;
  last_time_unreachable?: number;
}

export interface ServiceStatus extends StatusCommon {
  hostname: string;
  service_description: string;
  last_time_ok?: number;
  last_time_warning?: number;
  last_time_critical?: number;
  last_time_unknown?: number;
}

export interface Problem {
  kind: Kind;
  hostname: string;
  service_description?: string;
  state: number;
  state_text: string;
  is_hard_state: boolean;
  output: string;
  current_check_attempt: number;
  max_check_attempts: number;
  last_check: number;
  last_state_change: number;
  acknowledged: boolean;
  in_downtime: boolean;
  is_flapping: boolean;
  notifications_enabled: boolean;
  /** The host this service runs on is itself down, so this is probably a
   *  symptom rather than a separate incident. */
  host_is_down?: boolean;
}

export interface Downtime {
  kind: Kind;
  hostname: string;
  service_description?: string;
  internal_id: number;
  node_name?: string;
  author: string;
  comment: string;
  entry_time: number;
  scheduled_start_time: number;
  scheduled_end_time: number;
  is_fixed: boolean;
  duration: number;
  was_started: boolean;
  actual_start_time: number;
  triggered_by_id?: number;
  /** History only. */
  actual_end_time?: number;
  was_cancelled?: boolean;
}

export interface Acknowledgement {
  kind: Kind;
  hostname: string;
  service_description?: string;
  entry_time: number;
  state: number;
  state_text: string;
  author: string;
  comment: string;
  is_sticky: boolean;
  persistent_comment: boolean;
  notify_contacts: boolean;
  acknowledgement_type: number;
}

export interface LogEntry {
  id: number;
  entry_time: number;
  logentry_type: number;
  logentry_data: string;
  node_name?: string;
}

/**
 * What a detail page gets: the object, plus the records that explain
 * why it is quiet. Composed by the server so the page is one request
 * and stays one request on every live refresh.
 */
export interface ObjectContext {
  /** The windows the core is holding on this object right now. More
   *  than one can overlap, and the object stays in a downtime until the
   *  last of them ends. */
  downtimes: Downtime[];
  /** The acknowledgement in force, when there is one. */
  acknowledgement?: Acknowledgement;
}

export type HostDetail = HostStatus & ObjectContext;
export type ServiceDetail = ServiceStatus & ObjectContext;

export interface StateCounts {
  total: number;
  pending: number;
  by_state: Record<string, number>;
  problems: number;
  unhandled: number;
  acknowledged: number;
  in_downtime: number;
  flapping: number;
  notifications_disabled: number;
  active_checks_disabled: number;
}

export interface MonitoringNode {
  name: string;
  hosts: number;
  services: number;
  last_update: number;
}

export interface Summary {
  hosts: StateCounts;
  services: StateCounts;
  /** Newest status_update_time across both tables: how current this is. */
  last_update: number;
  nodes: MonitoringNode[];
  window: SummaryWindow;
}

/** One hour of the alert trend. Empty hours are present with a zero. */
export interface HourBucket {
  t: number;
  count: number;
}

/** The longest-running problem nobody has taken on. */
export interface OldestProblem {
  kind: Kind;
  hostname: string;
  service_description?: string;
  state: number;
  state_text: string;
  since: number;
}

/**
 * The recent past in numbers.
 *
 * `*_changed` counts objects that moved at all inside the window, not
 * how often: an object that flapped forty times counts once. Counting
 * the changes themselves would mean scanning a history table with no
 * index on its time column.
 */
export interface SummaryWindow {
  hours: number;
  since: number;
  hosts_changed: number;
  services_changed: number;
  notifications: number;
  notifications_by_hour: HourBucket[];
  /** The same count over the window before this one. */
  notifications_previous: number;
  oldest_problem?: OldestProblem;
}

// --- metrics ---------------------------------------------------------------

/** One measurable series on a service, and the window it covers. */
export interface MetricMeta {
  hostname: string;
  service_description: string;
  label: string;
  unit: string;
  first_seen: number;
  last_seen: number;
}

/**
 * One downsampled bucket. `min` and `max` travel with `avg` because
 * averaging a bucket hides the spike that caused the alert.
 */
export interface MetricPoint {
  t: number;
  avg: number;
  min: number;
  max: number;
}

export interface Series {
  label: string;
  unit: string;
  points: MetricPoint[];
}

export interface MetricResult {
  series: Series[];
  /** The resolution the provider settled on, so the chart can say
   *  "5-minute average" rather than implying raw samples. */
  bucket_seconds: number;
  from: number;
  to: number;
  /** Which provider answered: mysql today, graphite later. */
  source: string;
}

// --- history ---------------------------------------------------------------

export interface CheckResult {
  kind: Kind;
  hostname: string;
  service_description?: string;
  start_time: number;
  end_time: number;
  state: number;
  state_text: string;
  is_hard_state: boolean;
  output: string;
  long_output?: string;
  perfdata?: string;
  command?: string;
  current_check_attempt: number;
  max_check_attempts: number;
  latency: number;
  execution_time: number;
  timeout: number;
  early_timeout: boolean;
}

export interface StateChange {
  kind: Kind;
  hostname: string;
  service_description?: string;
  state_time: number;
  state: number;
  state_text: string;
  last_state: number;
  last_state_text: string;
  last_hard_state: number;
  is_hard_state: boolean;
  /** False for a row that repeats the previous state at a new check
   *  attempt rather than actually changing it. */
  is_transition: boolean;
  current_check_attempt: number;
  max_check_attempts: number;
  output: string;
  long_output?: string;
}

export interface NotificationRecord {
  kind: Kind;
  hostname: string;
  service_description?: string;
  start_time: number;
  end_time: number;
  contact_name: string;
  command_name: string;
  command_args?: string;
  state: number;
  state_text: string;
  reason_type: number;
  reason: string;
  output: string;
  ack_author?: string;
  ack_data?: string;
}

/** History lists add the window they cover and whether an object filter
 *  put them on the clustered key's fast path. */
export interface HistoryMeta extends ListMeta {
  from: number;
  to: number;
  scoped: boolean;
}

// --- command audit ---------------------------------------------------------

/** The actions the command log can hold, as the backend names them. */
export const COMMAND_ACTIONS = [
  'acknowledge',
  'remove_acknowledgement',
  'schedule_downtime',
  'delete_downtime',
  'reschedule',
  'submit_result',
  'custom_notification',
  'toggle',
] as const;

export type CommandAction = (typeof COMMAND_ACTIONS)[number];

/**
 * One submitted command, as the audit recorded it.
 *
 * One row per object: a downtime over forty services is forty records,
 * because "which of them did not take" is the question the log exists to
 * answer. `http_status` is the broker's answer and not Naemon's - 202
 * means accepted for delivery. Refusals are recorded too.
 */
export interface AuditRecord {
  id: number;
  username: string;
  action: string;
  /** `host` or `host/service`, as the command was addressed. */
  target: string;
  /** Whatever the request body held, minus nothing. */
  payload?: unknown;
  http_status: number;
  response: string;
  remote_ip?: string;
  created_at: number;
}
