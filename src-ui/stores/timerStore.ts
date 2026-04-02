import { create } from 'zustand';
import { invoke } from '@tauri-apps/api/core';
import type { Timer, CreateTimerRequest, UpdateTimerRequest } from '../types/index';

interface TimerState {
  timers: Timer[];
  activeTimer: Timer | null;
  isLoading: boolean;
  error: string | null;
  
  // Actions
  loadTimers: () => Promise<void>;
  createTimer: (req: CreateTimerRequest) => Promise<Timer | null>;
  updateTimer: (id: string, req: UpdateTimerRequest) => Promise<boolean>;
  deleteTimer: (id: string) => Promise<boolean>;
  startTimer: (id: string) => Promise<boolean>;
  pauseTimer: (id: string) => Promise<boolean>;
  resetTimer: (id: string) => Promise<boolean>;
  setActiveTimer: (timer: Timer | null) => void;
}

export const useTimerStore = create<TimerState>((set, get) => ({
  timers: [],
  activeTimer: null,
  isLoading: false,
  error: null,

  loadTimers: async () => {
    set({ isLoading: true, error: null });
    try {
      const timers = await invoke<Timer[]>('get_timers', { status: null });
      set({ timers, isLoading: false });
    } catch (err) {
      set({ error: String(err), isLoading: false });
    }
  },

  createTimer: async (req) => {
    try {
      const timer = await invoke<Timer>('create_timer', { req });
      set((state) => ({ timers: [timer, ...state.timers] }));
      return timer;
    } catch (err) {
      set({ error: String(err) });
      return null;
    }
  },

  updateTimer: async (id, req) => {
    try {
      const success = await invoke<boolean>('update_timer', { id, req });
      if (success) {
        await get().loadTimers();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  deleteTimer: async (id) => {
    try {
      const success = await invoke<boolean>('delete_timer', { id });
      if (success) {
        set((state) => ({
          timers: state.timers.filter((t) => t.id !== id),
          activeTimer: state.activeTimer?.id === id ? null : state.activeTimer,
        }));
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  startTimer: async (id) => {
    try {
      const success = await invoke<boolean>('start_timer', { id });
      if (success) {
        await get().loadTimers();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  pauseTimer: async (id) => {
    try {
      const success = await invoke<boolean>('pause_timer', { id });
      if (success) {
        await get().loadTimers();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  resetTimer: async (id) => {
    try {
      const success = await invoke<boolean>('reset_timer', { id });
      if (success) {
        await get().loadTimers();
      }
      return success;
    } catch (err) {
      set({ error: String(err) });
      return false;
    }
  },

  setActiveTimer: (timer) => set({ activeTimer: timer }),
}));
