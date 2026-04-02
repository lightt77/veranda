// Type definitions for Veranda frontend

export type TimerStatus = 'CREATED' | 'RUNNING' | 'PAUSED' | 'DONE';

export interface Timer {
  id: string;
  label: string;
  duration_seconds: number;
  remaining_seconds: number;
  status: TimerStatus;
  note?: string;
  started_at?: string;
  created_at: string;
}

export type StopwatchStatus = 'CREATED' | 'RUNNING' | 'PAUSED' | 'STOPPED';

export interface StopwatchLap {
  id: string;
  stopwatch_id: string;
  lap_number: number;
  lap_duration_ms: number;
  note?: string;
  created_at: string;
}

export interface Stopwatch {
  id: string;
  label: string;
  elapsed_seconds: number;
  status: StopwatchStatus;
  started_at?: string;
  created_at: string;
  laps?: StopwatchLap[];
}

export type TodoStatus = 'TODO' | 'IN_PROGRESS' | 'DONE' | 'ARCHIVED';

export interface Todo {
  id: string;
  task_name: string;
  task_desc?: string;
  project?: string;
  labels: string[];
  status: TodoStatus;
  last_status_update_ts?: string;
  created_at: string;
}

export interface ProjectSummary {
  name: string;
  todo_count: number;
  in_progress_count: number;
  done_count: number;
}

export interface LabelSummary {
  name: string;
  usage_count: number;
}

export interface CreateTimerRequest {
  label: string;
  duration_seconds: number;
  note?: string;
}

export interface UpdateTimerRequest {
  label?: string;
  note?: string;
}

export interface CreateStopwatchRequest {
  label: string;
}

export interface UpdateStopwatchRequest {
  label?: string;
}

export interface CreateTodoRequest {
  task_name: string;
  task_desc?: string;
  project?: string;
  labels?: string[];
}

export interface UpdateTodoRequest {
  task_name?: string;
  task_desc?: string;
  project?: string;
  labels?: string[];
}

export interface TodoFilter {
  status?: TodoStatus;
  project?: string;
  label?: string;
  search?: string;
}
