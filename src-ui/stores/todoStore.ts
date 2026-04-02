import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';
import type { 
  Todo, 
  ProjectSummary, 
  LabelSummary, 
  CreateTodoRequest, 
  UpdateTodoRequest,
  TodoFilter,
  TodoStatus 
} from '../types/index';

interface TodoState {
  todos: Todo[];
  projects: ProjectSummary[];
  labels: LabelSummary[];
  activeFilter: TodoFilter;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadTodos: (filter?: TodoFilter) => Promise<void>;
  loadProjects: () => Promise<void>;
  loadLabels: () => Promise<void>;
  createTodo: (req: CreateTodoRequest) => Promise<Todo | null>;
  updateTodo: (id: string, req: UpdateTodoRequest) => Promise<boolean>;
  deleteTodo: (id: string) => Promise<boolean>;
  moveTodo: (id: string, status: TodoStatus) => Promise<boolean>;
  setFilter: (filter: TodoFilter) => void;
}

export const useTodoStore = create<TodoState>((set, get) => ({
  todos: [],
  projects: [],
  labels: [],
  activeFilter: {},
  isLoading: false,
  error: null,

  loadTodos: async (filter) => {
    set({ isLoading: true, error: null });
    try {
      const todos = await invoke<Todo[]>('get_todos', { filter: filter || get().activeFilter });
      set({ todos, isLoading: false });
    } catch (err) {
      set({ error: String(err), isLoading: false });
    }
  },

  loadProjects: async () => {
    try {
      const projects = await invoke<ProjectSummary[]>('get_projects');
      set({ projects });
    } catch (err) {
      set({ error: String(err) });
    }
  },

  loadLabels: async () => {
    try {
      const labels = await invoke<LabelSummary[]>('get_labels');
      set({ labels });
    } catch (err) {
      set({ error: String(err) });
    }
  },

  createTodo: async (req) => {
    try {
      const todo = await invoke<Todo>('create_todo', { req });
      set((state) => ({ todos: [todo, ...state.todos] }));
      await get().loadProjects();
      return todo;
    } catch (err) {
      set({ error: String(err) });
      return null;
    }
  },

  updateTodo: async (id, req) => {
    try {
      const success = await invoke<boolean>('update_todo', { id, req });
      if (success) {
        await get().loadTodos();
        await get().loadProjects();
        await get().loadLabels();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  deleteTodo: async (id) => {
    try {
      const success = await invoke<boolean>('delete_todo', { id });
      if (success) {
        set((state) => ({
          todos: state.todos.filter((t) => t.id !== id),
        }));
        await get().loadProjects();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  moveTodo: async (id, status) => {
    try {
      const success = await invoke<boolean>('move_todo', { id, status });
      if (success) {
        await get().loadTodos();
        await get().loadProjects();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  setFilter: (filter) => {
    set({ activeFilter: filter });
    get().loadTodos(filter);
  },
}));
